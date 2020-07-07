# Redis Sentinel Exporter for Prometheus

[![BuildStatus Widget]][BuildStatus Result]
[![codecov](https://codecov.io/gh/leominov/redis_sentinel_exporter/branch/master/graph/badge.svg)](https://codecov.io/gh/leominov/redis_sentinel_exporter)

[BuildStatus Result]: https://travis-ci.com/leominov/redis_sentinel_exporter
[BuildStatus Widget]: https://travis-ci.com/leominov/redis_sentinel_exporter.svg?branch=master

This is a simple server that scrapes Redis Sentinel stats and exports them via HTTP for Prometheus consumption.

## Configuration

* `-debug` – Output verbose debug information. (env `DEBUG`)
* `-log-format` – Log format, valid options are txt and json. (env `LOG_FORMAT`) (default `txt`)
* `-namespace` – Namespace for metrics. (env `NAMESPACE`) (default `redis_sentinel")
* `-sentinel.addr` – Redis Sentinel host:port. (env `SENTINEL_ADDR`) (default `redis://127.0.0.1:26379`)
* `-sentinel.password` – Redis Sentinel password (env `SENTINEL_PASSWORD`) (optional).
* `-version` – Prints version and exit.
* `-web.listen-address` – Address to listen on for web interface and telemetry. (env `LISTEN_ADDRESS`) (default `:9355`)
* `-web.telemetry-path` – Path under which to expose metrics. (env `TELEMETRY_PATH`) (default `/metrics`)

## Links

* [Binary](https://github.com/leominov/redis_sentinel_exporter/releases)
* [Docker Image](https://hub.docker.com/r/leominov/redis_sentinel_exporter)
* [Grafana Dashboard](https://grafana.com/dashboards/9570)
