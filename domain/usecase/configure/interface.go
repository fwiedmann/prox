package configure

import "context"

type UseCase interface {
	StartConfigure(ctx context.Context, errChan chan<- error)
}
