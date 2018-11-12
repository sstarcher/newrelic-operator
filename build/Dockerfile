FROM alpine:3.6

RUN apk --update add ca-certificates

USER nobody

ADD build/_output/bin/newrelic-operator /usr/local/bin/newrelic-operator
