# Redis Sentinel Exporter for Prometheus

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

## Metrics

* redis_sentinel_exporter_last_scrape_duration_seconds
* redis_sentinel_exporter_scrapes_total
* redis_sentinel_exporter_last_scrape_error
* redis_sentinel_masters
* redis_sentinel_tilt
* redis_sentinel_running_scripts
* redis_sentinel_scripts_queue_length
* redis_sentinel_master_status
* redis_sentinel_master_slaves
* redis_sentinel_master_sentinels
* redis_sentinel_info
* redis_sentinel_uptime_seconds
