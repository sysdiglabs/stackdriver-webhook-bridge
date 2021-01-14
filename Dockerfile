FROM golang:1.13.3 as builder

COPY . /go/src/github.com/sysdiglabs/stackdriver-webhook-bridge
WORKDIR /go/src/github.com/sysdiglabs/stackdriver-webhook-bridge

ENV GO111MODULE=on
RUN make binary

FROM alpine
COPY --from=builder /go/src/github.com/sysdiglabs/stackdriver-webhook-bridge/build/stackdriver-webhook-bridge /stackdriver-webhook-bridge

# Use an unprivileged user
USER 65535

ENTRYPOINT ["/stackdriver-webhook-bridge"]
