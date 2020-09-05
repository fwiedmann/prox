package route

import (
	"context"

	"github.com/fwiedmann/prox/domain/entity"
)

// Router defines an API which is able to lookup all stored routes in the repository.
type Router interface {
	ListRoutes(ctx context.Context) ([]Route, error)
}

// Configurator defines an API which is able to configure the stored route entities.
type Configurator interface {
	CreateRoute(ctx context.Context, r Route) error
	UpdateRoute(ctx context.Context, r Route) error
	DeleteRoute(ctx context.Context, id entity.ID) error
	ListRoutes(ctx context.Context) ([]Route, error)
}

type repository interface {
	Router
	Configurator
}

// Manager is the API to interact with the stored route entities in the given repository.
// The Manager is responsible to validate the input before it gets proceeded by the repository.
type Manager interface {
	repository
}
