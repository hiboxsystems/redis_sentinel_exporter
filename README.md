# Redis Sentinel Exporter

## Usage

```
Usage of exporter:
  -debug
    	Output verbose debug information
  -log-format string
    	Log format, valid options are txt and json (default "txt")
  -namespace string
    	Namespace for metrics (default "redis_sentinel")
  -sentinel.addr string
    	Redis Sentinel host:port (default "redis://127.0.0.1:26379")
  -web.listen-address string
    	Address to listen on for web interface and telemetry. (default ":9355")
  -web.telemetry-path string
    	Path under which to expose metrics. (default "/metrics")
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
