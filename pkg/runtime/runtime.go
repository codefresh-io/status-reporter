package runtime

import (
	"context"

	"github.com/codefresh-io/status-reporter/pkg/codefresh"
	"github.com/codefresh-io/status-reporter/pkg/logger"
	"k8s.io/client-go/rest"
)

type (
	Options struct {
		Config *rest.Config
		Client *codefresh.Codefresh
		Logger logger.Logger
	}

	Runtime interface {
		Watch(ctx context.Context, namespace, workflowID string) error
	}
)
