FROM golang:1.10.3 as builder

WORKDIR /go/src/github.com/sstarcher/newrelic-operator
COPY . /go/src/github.com/sstarcher/newrelic-operator

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go/bin/newrelic-operator /go/src/github.com/sstarcher/newrelic-operator/cmd/manager/main.go

FROM alpine:3.10
RUN apk --update add ca-certificates
RUN addgroup -S newrelic-operator && adduser -S -G newrelic-operator newrelic-operator
USER newrelic-operator
COPY --from=builder /go/bin/newrelic-operator /usr/local/bin/newrelic-operator

ENTRYPOINT ["newrelic-operator"]
