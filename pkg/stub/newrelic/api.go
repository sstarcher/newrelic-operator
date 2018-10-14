package newrelic

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/IBM/newrelic-cli/newrelic"
	"github.com/IBM/newrelic-cli/utils"
	log "github.com/sirupsen/logrus"
)

// Client for newrelic wrapper
type Client struct {
	*newrelic.Client
}

// NewClient creates a fresh newrelic client
func NewClient() (*Client, error) {
	client, err := utils.GetNewRelicClient()
	if err != nil {
		return nil, err
	}

	return &Client{
		client,
	}, nil
}

// CreateDashboard in newrelic
func (c *Client) CreateDashboard(ctx context.Context, dashboard string) (*int64, error) {
	rsp, data, err := c.Dashboards.Create(ctx, dashboard)
	if err != nil {
		return nil, fmt.Errorf("newrelic api error %v", err)
	} else if rsp.StatusCode != 200 {
		return nil, fmt.Errorf("newrelic api error %v", rsp)
	}

	var result newrelic.CreateDashboardResponse
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}

	return result.Dashboard.ID, nil
}

// UpdateDashboard in newrelic
func (c *Client) UpdateDashboard(ctx context.Context, dashboard string, id int64) error {
	rsp, data, err := c.Dashboards.Update(ctx, dashboard, id)
	err = handleError(rsp, err)
	if err != nil {
		return err
	}

	log.Printf("%s", data)
	return nil
}

// DeleteDashboard  in newrelic
func (c *Client) DeleteDashboard(ctx context.Context, id int64) error {
	rsp, _, err := c.Dashboards.DeleteByID(ctx, id)
	err = handleError(rsp, err)
	if err != nil {
		return err
	}

	return nil
}

// func (c *Client) ReconcileDashboards(ctx context.Context) error {
// 	_, err := c.getDashboards(ctx)
// 	if err != nil {
// 		return nil
// 	}

// 	return nil
// }

func (c *Client) ListDashboards(ctx context.Context) ([]*newrelic.Dashboard, error) {
	rsp, data, err := c.Dashboards.ListAll(ctx, nil)
	err = handleError(rsp, err)
	if err != nil {
		return nil, err
	}

	var list newrelic.DashboardList
	err = json.Unmarshal(data, &list)
	if err != nil {
		return nil, err
	}

	return list.Dashboards, nil
}

func handleError(rsp *newrelic.Response, err error) error {
	if err != nil {
		return fmt.Errorf("newrelic api error %v", err)
	} else if rsp.StatusCode != 200 {
		return fmt.Errorf("newrelic api error %v", rsp)
	}
	return nil
}
