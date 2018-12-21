FROM golang:1.11.4 as builder
WORKDIR /go/src/github.com/leominov/redis_sentinel_exporter
COPY . .
RUN make


FROM scratch
COPY --from=builder /go/src/github.com/leominov/redis_sentinel_exporter/redis_sentinel_exporter /go/bin/redis_sentinel_exporter
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/go/bin/redis_sentinel_exporter"]  
