# NewRelic Kubernetes Operator

[![](https://images.microbadger.com/badges/image/sstarcher/newrelic-operator.svg)](http://microbadger.com/images/sstarcher/newrelic-operator "Get your own image badge on microbadger.com")
[![Docker Registry](https://img.shields.io/docker/pulls/sstarcher/newrelic-operator.svg)](https://registry.hub.docker.com/u/sstarcher/newrelic-operator)&nbsp;

__pre-alpha__ This is a work in progress to use the [operator-framework](https://github.com/operator-framework/operator-sdk) to create a controller and CRDs for NewRelic.  This allows us to create New Relic resources when creating  our services such as dashboards or synthetics.

# Capabilities

## Dashboards
* Can be created/updated/deleted
* Only the raw JSON for the dashboard is supported
* [Example](./examples/dashboard.yaml)

## Alert Channel
* Can be created/updated/deleted
* [Example](./examples/alert_channel.yaml)

## Alert Policy
* Can be created/updated/deleted
* Channels supported
* [Example](./examples/alert_policy.yaml)

## Monitor (Synthetics)
* Can be created/updated/deleted
* Can be tied to a policy
* [Example](./examples/monitor.yaml)


# Installation
* A helm chart is available in this [repository](./helm/newrelic-operator).
* To run the environment variable `NEW_RELIC_APIKEY` is required


## Todo
* Validate resources prior to calling API
* Need to support secret information like slack configuration and the ability to refer and re-use
