package v1alpha1

import (
	"os"

	nr "github.com/newrelic/newrelic-client-go/newrelic"
)

var client *nr.NewRelic

func init() {
	var err error
	// New Golang Client
	apiKey := os.Getenv("NEW_RELIC_APIKEY")
	client, err = nr.New(nr.ConfigAdminAPIKey(apiKey))
	if err != nil {
		panic(err)
	}
}
