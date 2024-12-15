package test

import (
	"sigs.k8s.io/e2e-framework/pkg/features"
	"testing"
)

func TestHttpEchoServer(t *testing.T) {
	install := features.
		New("TODO! Implement test for the echo server").
		Feature()

	ciTestEnv.Test(t, install)
}
