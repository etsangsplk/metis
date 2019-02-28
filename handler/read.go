package handler

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/digitalocean/metis/log"
	"github.com/golang/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/prometheus/prometheus/prompb"
	"github.com/prometheus/tsdb"
	"github.com/prometheus/tsdb/labels"
)

// RemoteRead is a public HTTP handler
func RemoteRead(db *tsdb.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query, err := parseRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		matchers := toMatchers(query)
		if len(matchers) == 0 {
			http.Error(w, "missing query matcher", http.StatusBadRequest)
			return
		}

		q, err := db.Querier(query.GetStartTimestampMs(), query.GetEndTimestampMs())
		if err != nil {
			log.Error("failed to create querier: %+v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer q.Close()

		set, err := q.Select(matchers...)
		if err != nil {
			log.Error("failed to execute query: %+v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		ts := toTimeseries(set)

		if err := writeResponse(w, ts); err != nil {
			log.Error("failed to write response: %+v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func parseRequest(r *http.Request) (*prompb.Query, error) {
	compressed, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %+v", err.Error())
	}

	buf, err := snappy.Decode(nil, compressed)
	if err != nil {
		return nil, fmt.Errorf("snappy decode failed: %+v", err)
	}

	var req prompb.ReadRequest
	if err := proto.Unmarshal(buf, &req); err != nil {
		return nil, fmt.Errorf("failed to unmarshal proto: %+v", err)
	}

	if len(req.GetQueries()) != 1 {
		return nil, fmt.Errorf("exactly one query must be sent. Got %d", len(req.GetQueries()))
	}

	return req.GetQueries()[0], nil
}

func toPbLabels(in labels.Labels) []*prompb.Label {
	res := make([]*prompb.Label, len(in))
	for i, l := range in {
		res[i] = &prompb.Label{
			Name:  l.Name,
			Value: l.Value,
		}
	}
	return res
}

func toMatchers(query *prompb.Query) []labels.Matcher {
	ms := query.GetMatchers()
	if ms == nil {
		return nil
	}

	mt := make([]labels.Matcher, len(ms))
	for i, m := range ms {
		switch m.GetType() {
		case prompb.LabelMatcher_EQ:
			mt[i] = labels.NewEqualMatcher(m.GetName(), m.GetValue())
		case prompb.LabelMatcher_NEQ:
			mt[i] = labels.Not(labels.NewEqualMatcher(m.GetName(), m.GetValue()))
		case prompb.LabelMatcher_RE:
			mt[i] = labels.NewMustRegexpMatcher(m.GetName(), m.GetValue())
		case prompb.LabelMatcher_NRE:
			mt[i] = labels.Not(labels.NewMustRegexpMatcher(m.GetName(), m.GetValue()))
		default:
			continue
		}
	}
	return mt

}

func toTimeseries(set tsdb.SeriesSet) []*prompb.TimeSeries {
	ts := make([]*prompb.TimeSeries, 0)

	for set.Next() {
		series := set.At()
		res := &prompb.TimeSeries{
			Labels:  toPbLabels(series.Labels()),
			Samples: []prompb.Sample{},
		}
		it := series.Iterator()
		for it.Next() {
			t, v := it.At()
			res.Samples = append(res.Samples, prompb.Sample{
				Timestamp: t,
				Value:     v,
			})
		}

		ts = append(ts, res)
	}
	return ts
}

func writeResponse(w http.ResponseWriter, ts []*prompb.TimeSeries) error {
	marshaled, err := proto.Marshal(&prompb.ReadResponse{
		Results: []*prompb.QueryResult{{Timeseries: ts}},
	})
	if err != nil {
		return fmt.Errorf("failed to marshal repsonse: %+v", err)
	}

	enc := snappy.Encode(nil, marshaled)

	if _, err := w.Write(enc); err != nil {
		return fmt.Errorf("failed to write response body: %+v", err)
	}
	return nil
}
