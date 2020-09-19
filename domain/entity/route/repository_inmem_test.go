package route

import (
	"context"
	"reflect"
	"sync"
	"testing"
)

func TestMemoryRepo_CreateRoute(t *testing.T) {
	t.Parallel()
	type fields struct {
		routes map[NameID]*Route
		mtx    sync.RWMutex
	}
	type args struct {
		ctx context.Context
		r   Route
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{

		{
			name: "ValidCreate",
			fields: fields{
				routes: make(map[NameID]*Route),
			},
			args: args{
				ctx: context.Background(),
				r: Route{
					NameID: "1",
				},
			},
			wantErr: false,
		},
		{
			name: "RouteAlreadyExists",
			fields: fields{
				routes: map[NameID]*Route{"1": {NameID: "1"}},
			},
			args: args{
				ctx: context.Background(),
				r: Route{
					NameID: "1",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests { //nolint
		t.Run(tt.name, func(t *testing.T) {
			m := &MemoryRepo{
				routes: tt.fields.routes,
				mtx:    tt.fields.mtx, //nolint
			}

			routeCountBeforeCreate := len(tt.fields.routes)
			err := m.CreateRoute(tt.args.ctx, &tt.args.r)
			if err != nil != tt.wantErr {
				t.Errorf("CreateRoute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && routeCountBeforeCreate != len(tt.fields.routes) {
				t.Errorf("CreateRoute() should not update routes count: want %d, got %d", routeCountBeforeCreate, len(tt.fields.routes))
				return
			}

			if err == nil && len(tt.fields.routes) == routeCountBeforeCreate {
				t.Errorf("CreateRoute() should update routes count: want %d, got %d", routeCountBeforeCreate, len(tt.fields.routes))
				return
			}

		})
	}
}

func TestMemoryRepo_UpdateRoute(t *testing.T) {
	t.Parallel()
	type fields struct {
		routes map[NameID]*Route
		mtx    sync.RWMutex
	}
	type args struct {
		ctx context.Context
		r   Route
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "ValidUpdate",
			fields: fields{
				routes: map[NameID]*Route{"1": {
					NameID:   "1",
					Priority: 1,
				}},
			},
			args: args{
				ctx: context.Background(),
				r: Route{
					NameID:   "1",
					Priority: 2,
				},
			},
			wantErr: false,
		},
		{
			name: "RouteNotFound",
			fields: fields{
				routes: map[NameID]*Route{"1": {
					NameID:   "1",
					Priority: 1,
				}},
			},
			args: args{
				ctx: context.Background(),
				r: Route{
					NameID:   "2",
					Priority: 2,
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests { //nolint
		t.Run(tt.name, func(t *testing.T) {
			m := &MemoryRepo{
				routes: tt.fields.routes,
				mtx:    tt.fields.mtx, //nolint
			}

			err := m.UpdateRoute(tt.args.ctx, &tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateRoute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if !reflect.DeepEqual(tt.args.r, *tt.fields.routes[tt.args.r.NameID]) {
					t.Errorf("UpdateRoute() did not update the route: want %+v, got %+v", tt.args.r, tt.fields.routes[tt.args.r.NameID])
					return
				}
			}

		})
	}
}

func TestMemoryRepo_DeleteRoute(t *testing.T) {
	t.Parallel()
	type fields struct {
		routes map[NameID]*Route
		mtx    sync.RWMutex
	}
	type args struct {
		ctx context.Context
		id  NameID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "ValidDelete",
			fields: fields{
				routes: map[NameID]*Route{"1": {NameID: "1"}},
			},
			args: args{
				ctx: context.Background(),
				id:  "1",
			},
			wantErr: false,
		},
		{
			name: "RouteNotFound",
			fields: fields{
				routes: map[NameID]*Route{"1": {NameID: "1"}},
			},
			args: args{
				ctx: context.Background(),
				id:  "2",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests { //nolint
		t.Run(tt.name, func(t *testing.T) {
			m := &MemoryRepo{
				routes: tt.fields.routes,
				mtx:    tt.fields.mtx, //nolint
			}
			err := m.DeleteRoute(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteRoute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err == nil {
				if _, ok := tt.fields.routes[tt.args.id]; ok {
					t.Errorf("DeleteRoute() should delete route with ID %s but is still present", tt.args.id)
					return
				}
			}

		})
	}
}

func TestMemoryRepo_ListRoutes(t *testing.T) {
	t.Parallel()
	type fields struct {
		routes map[NameID]*Route
		mtx    sync.RWMutex
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{

		{
			name: "ValidList",
			fields: fields{
				routes: map[NameID]*Route{"1": {}, "2": {}},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests { //nolint
		t.Run(tt.name, func(t *testing.T) {
			m := &MemoryRepo{
				routes: tt.fields.routes,
				mtx:    tt.fields.mtx, //nolint
			}
			got, err := m.ListRoutes(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListRoutes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && len(got) != len(tt.fields.routes) {
				t.Errorf("ListRoutes() returned routes size differs with the stored size: got %d, want %d", len(got), len(tt.fields.routes))
				return
			}
		})
	}
}
