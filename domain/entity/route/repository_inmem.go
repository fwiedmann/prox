package route

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrorAlreadyExists = errors.New("route already exists")
	ErrorNotFound      = errors.New("route not found")
)

// MemoryRepo implements the repository interface. All data will be stored in the memory.
type MemoryRepo struct {
	routes map[NameID]*Route
	mtx    sync.RWMutex
}

// NewInMemRepo initialize a empty MemoryRepo which implements the repository interface.
func NewInMemRepo() *MemoryRepo {
	return &MemoryRepo{
		routes: make(map[NameID]*Route),
		mtx:    sync.RWMutex{},
	}
}

// CreateRoute stores a new route in the repository.routes. If route already exists it returns an ErrorAlreadyExists.
func (m *MemoryRepo) CreateRoute(ctx context.Context, r *Route) error {
	m.mtx.RLock()
	if _, ok := m.routes[r.NameID]; ok {
		m.mtx.RUnlock()
		return ErrorAlreadyExists
	}
	m.mtx.RUnlock()

	m.mtx.Lock()
	m.routes[r.NameID] = r
	m.mtx.Unlock()
	return nil
}

// UpdateRoute with the given route in the repository.routes. If the given route does not exists it returns an ErrorNotFound.
func (m *MemoryRepo) UpdateRoute(ctx context.Context, r *Route) error {
	m.mtx.RLock()
	if _, ok := m.routes[r.NameID]; !ok {
		m.mtx.RUnlock()
		return ErrorNotFound
	}
	m.mtx.RUnlock()

	m.mtx.Lock()
	m.routes[r.NameID] = r
	m.mtx.Unlock()
	return nil
}

// DeleteRoute with the given entity.id in the repository.routes. If the given route does not exists it returns an ErrorNotFound.
func (m *MemoryRepo) DeleteRoute(_ context.Context, id NameID) error {
	m.mtx.RLock()
	if _, ok := m.routes[id]; !ok {
		m.mtx.RUnlock()
		return ErrorNotFound
	}
	m.mtx.RUnlock()

	m.mtx.Lock()
	delete(m.routes, id)
	m.mtx.Unlock()
	return nil
}

// ListRoutes which are stored in the repository.routes
func (m *MemoryRepo) ListRoutes(ctx context.Context) ([]*Route, error) {
	routes := make([]*Route, 0)
	m.mtx.RLock()
	for _, r := range m.routes {
		routes = append(routes, r)
	}
	m.mtx.RUnlock()
	return routes, nil
}
