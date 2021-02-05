package kubemq

import (
	"context"
	"fmt"
	"github.com/kubemq-io/file-uploader/config"
	"github.com/kubemq-io/file-uploader/pkg/uuid"
	"github.com/kubemq-io/file-uploader/types"
	"github.com/kubemq-io/kubemq-go"
	"strconv"
	"strings"
)

type Client struct {
	cfg    *config.Config
	client *kubemq.Client
}

func NewClient() *Client {
	return &Client{}

}
func parseAddress(address string) (string, int, error) {
	var host string
	var port int
	hostPort := strings.Split(address, ":")

	if len(hostPort) >= 1 {
		host = hostPort[0]
	}
	if len(hostPort) >= 2 {
		port, _ = strconv.Atoi(hostPort[1])
	}
	if host == "" {
		return "", 0, fmt.Errorf("no valid host found")
	}
	if port < 0 {
		return "", 0, fmt.Errorf("no valid port found")
	}
	return host, port, nil
}
func (c *Client) Init(ctx context.Context, cfg *config.Config) error {
	c.cfg = cfg
	host, port, err := parseAddress(cfg.Target.Address)
	if err != nil {
		return err
	}
	if cfg.Target.ClientId == "" {
		cfg.Target.ClientId = uuid.New().String()
	}
	c.client, err = kubemq.NewClient(ctx,
		kubemq.WithAddress(host, port),
		kubemq.WithTransportType(kubemq.TransportTypeGRPC),
		kubemq.WithClientId(cfg.Target.ClientId),
		kubemq.WithAuthToken(cfg.Target.AuthToken),
		kubemq.WithCheckConnection(false),
	)
	if err != nil {
		return err
	}
	return nil
}
func (c *Client) Send(ctx context.Context, req *types.Request) (*types.Response, error) {

	resp, err := c.client.SendQueueMessage(ctx, req.ToQueueMessage().SetChannel(c.cfg.Target.Queue))
	if err != nil {
		return nil, err
	}
	if resp.IsError {
		return types.NewResponse().SetError(fmt.Errorf(resp.Error)), nil
	}
	return types.NewResponse(), nil
}
