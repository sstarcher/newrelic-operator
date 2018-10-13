FROM alpine:3.6

RUN adduser -D newrelic-operator
USER newrelic-operator

ADD tmp/_output/bin/newrelic-operator /usr/local/bin/newrelic-operator
