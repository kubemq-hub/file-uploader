package null

import (
	"context"
	"github.com/kubemq-io/file-uploader/pkg/logger"
	"github.com/kubemq-io/file-uploader/types"
	"time"
)

type Null struct {
	logger *logger.Logger
}

func NewNullSenders() *Null {
	return &Null{
		logger: logger.NewLogger("null-sender"),
	}
}

func (n *Null) Send(ctx context.Context, req *types.Request) (*types.Response, error) {
	n.logger.Infof("sending file: %s", req.Metadata.String())
	time.Sleep(2 * time.Second)
	return types.NewResponse().SetMetadata(req.Metadata), nil
}
