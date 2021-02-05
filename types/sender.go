package types

import "context"

type Sender interface {
	Send(ctx context.Context, request *Request) (*Response, error)
}
