package v1alpha1

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/IBM/newrelic-cli/newrelic"
	"github.com/IBM/newrelic-cli/utils"
)

var client *newrelic.Client
var clientSythetics *newrelic.Client

func init() {
	var err error
	client, err = utils.GetNewRelicClient()
	if err != nil {
		panic(err)
	}

	clientSythetics, err = utils.GetNewRelicClient("synthetics")
	if err != nil {
		panic(err)
	}

	clean := os.Getenv("NEW_RELIC_OPERATOR_CLEANUP")
	if clean != "" {
		list, err := listDashboards(context.Background())
		if err != nil {
			panic(err)
		}
		for _, item := range list {
			if item.OwnerEmail != nil && *item.OwnerEmail == clean {
				_, _, err := client.Dashboards.DeleteByID(context.Background(), *item.ID)
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

func handleErrorMessage(format string, rsp *newrelic.Response, err error) error {
	if err != nil {
		return fmt.Errorf(format, err)
	} else if rsp.StatusCode < 200 || rsp.StatusCode >= 300 {
		return fmt.Errorf(format, rsp)
	}
	return nil
}

func handleError(rsp *newrelic.Response, err error) error {
	if err != nil {
		return fmt.Errorf("newrelic api error %v", err)
	} else if rsp.StatusCode < 200 || rsp.StatusCode >= 300 {
		body := responseBodyToString(rsp)
		return fmt.Errorf("%s", body)
	}
	return nil
}

func responseBodyToString(rsp *newrelic.Response) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(rsp.Body)
	return buf.String()
}
