FROM golang:1.10.3 as builder

WORKDIR /go/src/github.com/sstarcher/newrelic-operator
COPY . /go/src/github.com/sstarcher/newrelic-operator

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /go/bin/newrelic-operator /go/src/github.com/sstarcher/newrelic-operator/cmd/manager/main.go

FROM alpine:3.6
RUN apk --update add ca-certificates
COPY --from=builder /go/bin/newrelic-operator /usr/local/bin/newrelic-operator

ENTRYPOINT ["newrelic-operator"]
