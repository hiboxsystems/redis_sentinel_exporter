# Redis Sentinel Exporter for Prometheus

[![BuildStatus Widget]][BuildStatus Result]
[![codecov](https://codecov.io/gh/leominov/redis_sentinel_exporter/branch/master/graph/badge.svg)](https://codecov.io/gh/leominov/redis_sentinel_exporter)

[BuildStatus Result]: https://travis-ci.com/leominov/redis_sentinel_exporter
[BuildStatus Widget]: https://travis-ci.com/leominov/redis_sentinel_exporter.svg?branch=master

This is a simple server that scrapes Redis Sentinel stats and exports them via HTTP for Prometheus consumption.

## Getting Started

To run it:

```
./redis_sentinel_exporter [flags]
```

Help on flags:

```
./redis_sentinel_exporter --help
```

## Links

* [Binary](https://github.com/leominov/redis_sentinel_exporter/releases)
* [Docker Image](https://hub.docker.com/r/leominov/redis_sentinel_exporter)
* [Grafana Dashboard](https://grafana.com/dashboards/9570)
