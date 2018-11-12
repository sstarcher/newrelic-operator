package e2e

import (
	"testing"

	"github.com/operator-framework/operator-sdk/pkg/test"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
)

func TestMain(m *testing.M) {
	test.MainEntry(m)
}
