package route

import (
	"context"
	"errors"

	"github.com/fwiedmann/prox/domain/entity"
)

var (
	NoEntityID = errors.New("route ID is empty")
)

type manager struct {
	repo repository
}

func (m *manager) UpdateRoute(ctx context.Context, r Route) error {
	if r.ID == "" {
		return NoEntityID
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}

	return m.repo.UpdateRoute(ctx, r)
}

func (m *manager) ListRoutes(ctx context.Context) ([]Route, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return m.repo.ListRoutes(ctx)
}

func (m *manager) CreateRoute(ctx context.Context, r Route) error {

	// Todo: validate route e.g check host,path and hostregex,pathRegex
	r.ID = entity.NewID()
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return m.repo.CreateRoute(ctx, r)
}

func (m *manager) DeleteRoute(ctx context.Context, id entity.ID) error {
	if id == "" {
		return NoEntityID
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}
	return m.repo.DeleteRoute(ctx, id)
}

func NewManager(r repository) Manager {
	return &manager{
		repo: r,
	}
}
