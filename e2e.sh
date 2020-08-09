#!/usr/bin/env bash

set -x
set -o pipefail
set -o nounset

version="$1"
skip_re="^(redis_sentinel_build_info|redis_sentinel_info|redis_sentinel_used_cpu|redis_sentinel_exporter_last_scrape_duration_seconds|redis_sentinel_uptime_in_seconds|redis_sentinel_connections_received_total|redis_sentinel_net|redis_sentinel_instantaneous|redis_sentinel_process_id)"

echo "==> Redis $version"

rm -rf "redis-${version}" "redis-${version}.tar.gz"

wget "http://download.redis.io/releases/redis-${version}.tar.gz"
tar -zxvf "redis-${version}.tar.gz"
cd "redis-${version}"
make

nohup ./src/redis-server &

for i in {1..5} ; do
  if ./src/redis-cli -p 6379 PING; then
    break
  fi
  sleep 1
done

cp ../test_data/sentinel.conf sentinel.conf
nohup ./src/redis-sentinel sentinel.conf &

for i in {1..5} ; do
  if ./src/redis-cli -p 26379 PING; then
    break
  fi
  sleep 1
done

cd ../

go build
nohup ./redis_sentinel_exporter &
wget --retry-connrefused --tries=5 -O - "127.0.0.1:9355/metrics"| grep "redis_" | grep -E -v "${skip_re}" > "e2e-output.txt"

diff -u \
  "test_data/e2e-output.txt" \
  "e2e-output.txt"
