# Redis Sentinel Exporter for Prometheus

This is a simple server that scrapes Redis Sentinel stats and exports them via HTTP for Prometheus consumption.

## Configuration

* `-debug` (env `DEBUG`) – Output verbose debug information.
* `-log-format` (env `LOG_FORMAT`) – Log format, valid options are txt and json. (default `txt`)
* `-namespace` (env `NAMESPACE`) – Namespace for metrics. (default `redis_sentinel`)
* `-sentinel.addr` (env `SENTINEL_ADDR`) – Redis Sentinel host:port. (default `redis://127.0.0.1:26379`)
* `-sentinel.password` (env `SENTINEL_PASSWORD`) – Redis Sentinel password (optional).
* `-sentinel.password-file` (env `SENTINEL_PASSWORD_FILE`) - Path to Redis Sentinel password file (optional).
* `-version` – Prints version and exit.
* `-web.listen-address` (env `LISTEN_ADDRESS`) – Address to listen on for web interface and telemetry. (default `:9355`)
* `-web.telemetry-path` (env `TELEMETRY_PATH`) – Path under which to expose metrics. (default `/metrics`)

## Links

* [Binary](https://github.com/leominov/redis_sentinel_exporter/releases)
* [Docker Image](https://hub.docker.com/r/leominov/redis_sentinel_exporter)
* [Grafana Dashboard](https://grafana.com/dashboards/9570)

## Forked from

* <https://github.com/hiboxsystems/redis_sentinel_exporter>
* <https://github.com/leominov/redis_sentinel_exporter>

## Docker

Images are pushed to <https://hub.docker.com/repository/docker/monotek/redis-sentinel-exporter>

Run via:

```bash
docker run -it --rm -p 9355:9355 monotek/redis-sentinel-exporter
```
