package test

import (
	"context"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"
	"testing"
)

func TestPrometheus(t *testing.T) {
	feature := features.
		New("Testing Prometheus deployment").
		Setup(func(ctx context.Context, t *testing.T, cfg *envconf.Config) context.Context {

			return ctx
		}).Feature()
	ciTestEnv.Test(t, feature)

}
