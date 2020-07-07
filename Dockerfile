FROM golang:1.14 as builder
WORKDIR /go/src/github.com/leominov/redis_sentinel_exporter
COPY . .
RUN make build

FROM scratch
COPY --from=builder /go/src/github.com/leominov/redis_sentinel_exporter/redis_sentinel_exporter /go/bin/redis_sentinel_exporter
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/go/bin/redis_sentinel_exporter"]  
