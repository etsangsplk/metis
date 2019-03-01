# Metis TSDB

[![Build
Status](https://travis-ci.org/digitalocean/metis.svg?branch=master)](https://travis-ci.org/digitalocean/metis)
[![Go Report Card](https://goreportcard.com/badge/github.com/digitalocean/metis)](https://goreportcard.com/report/github.com/digitalocean/metis)
[![Coverage Status](https://coveralls.io/repos/github/digitalocean/metis/badge.svg?branch=feat%2Fadd-coveralls-report)](https://coveralls.io/github/digitalocean/metis?branch=feat%2Fadd-coveralls-report)

```
Usage of metis:
  -data-dir string
        filesystem path to store tsdb data (default "./data")
  -listen-addr string
        web address to listen for RemoteRead and RemoteWrite (default ":8080")
  -log-level string
        log level info or error (default "info")
  -no-lockfile
        disable the use of the TSDB lockfile
  -retention-duration duration
        tsdb retention time (default 48h0m0s)
  -version
        show version information
  -wal-segment-size string
        Write Ahead Log segment size (default "128MB")
```

Metis TSDB is [TSDB](https://github.com/prometheus/tsdb) wrapped with HTTP
endpoints for RemoteRead and RemoteWrite. You can run `metis` on a persistent
disk and run an ephemeral instance of Prometheus (i.e. in Kubernetes),
configured to remote-read and remote-write against `metis`. You can also write
directly to the TSDB even while Prometheus is connected to it.

To see an example, see the [docker-compose.yml](docker-compose.yml) file.

See [the Prometheus docs](https://prometheus.io/docs/prometheus/latest/storage/#remote-storage-integrations) for more details
