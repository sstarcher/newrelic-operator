package newrelic

import (
	"context"
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	c, err := NewClient()
	if err != nil {
		t.Error(err)
	}

	id, err := c.FindDashboard(context.Background(), "Test Shane")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(*id)

}

var dat = `{
	"dashboard": {
	  "title": "Test MTB Dashboard-Auto",
	  "description": null,
	  "icon": "bar-chart",
	  "visibility": "all",
	  "editable": "editable_by_owner",
	  "metadata": {
		"version": 1
	  },
	  "widgets": []
	 }
  }`
