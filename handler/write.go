package handler

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	"github.com/prometheus/prometheus/prompb"
	"github.com/prometheus/tsdb"
	"github.com/prometheus/tsdb/labels"
)

// RemoteWrite is an HTTP handler to handle Prometheus remote_write
func RemoteWrite(db *tsdb.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		timeseries, err := readRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if len(timeseries) == 0 {
			// nothing was sent so just nop
			return
		}

		ap := db.Appender()
		defer ap.Commit()
		for _, ts := range timeseries {
			lbls := make(labels.Labels, len(ts.Labels))
			for i, l := range ts.Labels {
				lbls[i] = labels.Label{
					Name:  l.GetName(),
					Value: l.GetValue(),
				}
			}

			var ref uint64
			var err error
			for _, s := range ts.Samples {
				if ref == 0 {
					ref, err = ap.Add(lbls, s.GetTimestamp(), s.GetValue())
				} else {
					err = ap.AddFast(ref, s.GetTimestamp(), s.GetValue())
				}
				if err != nil {
					log.Printf("failed writing sample to store: %+v\n", err)
				}
			}
		}

	}
}

func readRequest(r *http.Request) ([]*prompb.TimeSeries, error) {
	compressed, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %+v", err)
	}

	reqBuf, err := snappy.Decode(nil, compressed)
	if err != nil {
		return nil, fmt.Errorf("failed to snappy.Decode: %+v", err)
	}

	var req prompb.WriteRequest
	if err := proto.Unmarshal(reqBuf, &req); err != nil {
		return nil, fmt.Errorf("failed to proto.Unmarshal: %+v", err)
	}
	return req.GetTimeseries(), nil
}
