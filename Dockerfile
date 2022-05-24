FROM golang:1.18 as builder
WORKDIR /redis-sentinel-exporter
COPY . .
RUN ./build.sh

FROM scratch
COPY --from=builder /redis-sentinel-exporter/redis_sentinel_exporter /usr/local/bin/redis_sentinel_exporter
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER nobody

ENTRYPOINT ["/usr/local/bin/redis_sentinel_exporter"]
