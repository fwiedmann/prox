package route

import (
	"context"
	"errors"
	"sync"

	"github.com/fwiedmann/prox/domain/entity"
)

var (
	AlreadyExistsError = errors.New("route already exists")
	NotFoundError      = errors.New("route not found")
)

// MemoryRepo implements the repository interface. All data will be stored in the memory.
type MemoryRepo struct {
	routes map[entity.ID]Route
	mtx    sync.RWMutex
}

// NewInMemRepo initialize a empty MemoryRepo which implements the repository interface.
func NewInMemRepo() *MemoryRepo {
	return &MemoryRepo{
		routes: make(map[entity.ID]Route),
		mtx:    sync.RWMutex{},
	}
}

// CreateRoute stores a new route in the repository.routes. If route already exists it returns an AlreadyExistsError.
func (m *MemoryRepo) CreateRoute(_ context.Context, r Route) error {
	m.mtx.RLock()
	if _, ok := m.routes[r.ID]; ok {
		return AlreadyExistsError
		m.mtx.RUnlock()
	}
	m.mtx.RUnlock()

	m.mtx.Lock()
	m.routes[r.ID] = r
	m.mtx.Unlock()
	return nil
}

// UpdateRoute with the given route in the repository.routes. If the given route does not exists it returns an NotFoundError.
func (m *MemoryRepo) UpdateRoute(_ context.Context, r Route) error {
	m.mtx.RLock()
	if _, ok := m.routes[r.ID]; !ok {
		return NotFoundError
		m.mtx.RUnlock()
	}
	m.mtx.RUnlock()

	m.mtx.Lock()
	m.routes[r.ID] = r
	m.mtx.Unlock()
	return nil
}

// DeleteRoute with the given entity.id in the repository.routes. If the given route does not exists it returns an NotFoundError.
func (m *MemoryRepo) DeleteRoute(_ context.Context, id entity.ID) error {
	m.mtx.RLock()
	if _, ok := m.routes[id]; !ok {
		return NotFoundError
		m.mtx.RUnlock()
	}
	m.mtx.RUnlock()

	m.mtx.Lock()
	delete(m.routes, id)
	m.mtx.Unlock()
	return nil
}

// ListRoutes which are stored in the repository.routes
func (m *MemoryRepo) ListRoutes(_ context.Context) ([]Route, error) {
	routes := make([]Route, 0)
	m.mtx.RLock()
	for _, r := range m.routes {
		routes = append(routes, r)
	}
	m.mtx.RUnlock()
	return routes, nil
}
