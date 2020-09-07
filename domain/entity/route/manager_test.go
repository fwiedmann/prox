package route

import (
	"context"
	"errors"
	"testing"
)

func Test_manager_CreateRoute(t *testing.T) {
	t.Parallel()
	type fields struct {
		repo repository
	}
	type args struct {
		ctx context.Context
		r   *Route
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantErr   bool
		errType   error
		cancelCtx bool
	}{
		{
			name:   "InvalidCacheDuration",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				r: &Route{
					NameID:               "test-route",
					Hostname:             "docker.com",
					CacheTimeOutDuration: "10seconds",
				},
			},
			wantErr: true,
			errType: ErrorInvalidCacheTimeOutDuration,
		},
		{
			name:   "InvalidUpstreamTimeOutDuration",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				r: &Route{
					NameID:                  "test-route",
					Hostname:                "docker.com",
					UpstreamTimeoutDuration: "10seconds",
				},
			},
			wantErr: true,
			errType: ErrorInvalidUpstreamTimeOutDuration,
		},
		{
			name:   "ErrorEmptyRoute",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				r:   nil,
			},
			wantErr: true,
			errType: ErrorEmptyRoute,
		},
		{
			name:   "NoEntityNameID",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				r:   &Route{},
			},
			wantErr: true,
			errType: ErrorNoEntityID,
		},
		{
			name:   "EmptyRequestIdentifiers",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				r: &Route{
					NameID: "test-route",
				},
			},
			wantErr: true,
			errType: ErrorEmptyRequestIdentifiers,
		},
		{
			name:   "ErrorDuplicatedRequestIdentifier",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				r: &Route{
					NameID:         "test-route",
					Hostname:       "docker.com",
					HostnameRegexp: "^docker.com$",
					Path:           "/hello",
					PathRegexp:     "[\\s\\S]*",
				},
			},
			wantErr: true,
			errType: ErrorDuplicatedRequestIdentifier,
		},
		{
			name:   "ErrorDuplicatedHostRequestIdentifier",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				r: &Route{
					NameID:         "test-route",
					Hostname:       "docker.com",
					HostnameRegexp: "^docker.com$",
				},
			},
			wantErr: true,
			errType: ErrorDuplicatedHostRequestIdentifier,
		},
		{
			name:   "ErrorDuplicatedPathRequestIdentifier",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				r: &Route{
					NameID:     "test-route",
					Path:       "/hello",
					PathRegexp: "[\\s\\S]*",
				},
			},
			wantErr: true,
			errType: ErrorDuplicatedPathRequestIdentifier,
		},
		{
			name:   "ErrorInvalidHostName",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				r: &Route{
					NameID:   "test-route",
					Hostname: "docker!!!.com",
				},
			},
			wantErr: true,
			errType: ErrorInvalidHostName,
		},
		{
			name: "ValidHostNameAndPath",
			fields: fields{
				repo: NewInMemRepo(),
			},
			args: args{
				ctx: context.Background(),
				r: &Route{
					NameID:   "test-route",
					Hostname: "docker.com",
					Path:     "/test",
				},
			},
			wantErr: false,
			errType: nil,
		},
		{
			name: "ValidHostNameAndPathExpr",
			fields: fields{
				NewInMemRepo(),
			},
			args: args{
				ctx: context.Background(),
				r: &Route{
					NameID:         "test-route",
					HostnameRegexp: "docker.com",
					PathRegexp:     "/test",
				},
			},
			wantErr: false,
			errType: nil,
		},
		{
			name: "ValidHostnameWithWildcardPath",
			fields: fields{
				NewInMemRepo(),
			},
			args: args{
				ctx: context.Background(),
				r: &Route{
					NameID:         "test-route",
					HostnameRegexp: "docker.com",
				},
			},
			wantErr: false,
			errType: nil,
		},
		{
			name: "ValidPathWithWildcardHostname",
			fields: fields{
				NewInMemRepo(),
			},
			args: args{
				ctx: context.Background(),
				r: &Route{
					NameID:         "test-route",
					HostnameRegexp: "docker.com",
				},
			},
			wantErr: false,
			errType: nil,
		},
		{
			name: "ValidPathWithWildcardHostnameWithNegativeMaxBodySizeInBytes",
			fields: fields{
				NewInMemRepo(),
			},
			args: args{
				ctx: context.Background(),
				r: &Route{
					NameID:                      "test-route",
					HostnameRegexp:              "docker.com",
					CacheMaxBodySizeInMegaBytes: -10,
				},
			},
			wantErr: false,
			errType: nil,
		},
		{
			name: "RepoError",
			fields: fields{
				repo: &MemoryRepo{routes: map[NameID]*Route{"test-route": {}}},
			},
			args: args{
				ctx: context.Background(),
				r: &Route{
					NameID:         "test-route",
					HostnameRegexp: "docker.com",
				},
			},
			wantErr: true,
			errType: ErrorAlreadyExists,
		},
		{
			name: "ContextCanceled",
			fields: fields{
				NewInMemRepo(),
			},
			args: args{
				ctx: context.Background(),
				r: &Route{
					NameID:         "test-route",
					HostnameRegexp: "docker.com",
				},
			},
			wantErr:   true,
			cancelCtx: true,
			errType:   context.Canceled,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				repo: tt.fields.repo,
			}
			c := tt.args.ctx
			cancel := func() {}
			if tt.cancelCtx {
				c, cancel = context.WithCancel(tt.args.ctx)
				cancel()
			}

			err := m.CreateRoute(c, tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateRoute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.wantErr {
				if !errors.Is(err, tt.errType) {
					t.Errorf("CreateRoute() returned error should be %s, but is %s", tt.errType, err)
				}
			}
		})
	}
}

func Test_manager_UpdateRoute(t *testing.T) {
	t.Parallel()
	type fields struct {
		repo repository
	}
	type args struct {
		ctx context.Context
		r   *Route
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantErr   bool
		errType   error
		cancelCtx bool
	}{
		{
			name:   "ErrorEmptyRoute",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				r:   nil,
			},
			wantErr: true,
			errType: ErrorEmptyRoute,
		},
		{
			name:   "NoEntityNameID",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				r:   &Route{},
			},
			wantErr: true,
			errType: ErrorNoEntityID,
		},
		{
			name: "RepoError",
			fields: fields{
				repo: &MemoryRepo{routes: map[NameID]*Route{}},
			},
			args: args{
				ctx: context.Background(),
				r: &Route{
					NameID:         "test-route",
					HostnameRegexp: "docker.com",
				},
			},
			wantErr: true,
			errType: ErrorNotFound,
		},
		{
			name:   "ContextCancelled",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				r:   &Route{NameID: "test-route"}},
			wantErr:   true,
			errType:   context.Canceled,
			cancelCtx: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				repo: tt.fields.repo,
			}

			c := tt.args.ctx
			cancel := func() {}
			if tt.cancelCtx {
				c, cancel = context.WithCancel(tt.args.ctx)
				cancel()
			}
			err := m.UpdateRoute(c, tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateRoute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.wantErr {
				if !errors.Is(err, tt.errType) {
					t.Errorf("CreateRoute() returned error should be %s, but is %s", tt.errType, err)
				}
			}
		})
	}
}

func Test_manager_ListRoutes(t *testing.T) {
	t.Parallel()
	type fields struct {
		repo repository
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantErr   bool
		cancelCtx bool
		errType   error
	}{
		{
			name:      "ContextCancelled",
			fields:    fields{},
			args:      args{ctx: context.Background()},
			wantErr:   true,
			errType:   context.Canceled,
			cancelCtx: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				repo: tt.fields.repo,
			}

			c := tt.args.ctx
			cancel := func() {}
			if tt.cancelCtx {
				c, cancel = context.WithCancel(tt.args.ctx)
				cancel()
			}
			_, err := m.ListRoutes(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListRoutes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr {
				if !errors.Is(err, tt.errType) {
					t.Errorf("CreateRoute() returned error should be %s, but is %s", tt.errType, err)
				}
			}
		})
	}
}

func Test_manager_DeleteRoute(t *testing.T) {
	t.Parallel()
	type fields struct {
		repo repository
	}
	type args struct {
		ctx context.Context
		id  NameID
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantErr   bool
		errType   error
		cancelCtx bool
	}{
		{
			name:   "ContextCancelled",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				id:  "test-route"},
			wantErr:   true,
			errType:   context.Canceled,
			cancelCtx: true,
		},
		{
			name:   "EmptyNameIDError",
			fields: fields{},
			args: args{
				ctx: context.Background(),
				id:  ""},
			wantErr:   true,
			errType:   ErrorNoEntityID,
			cancelCtx: false,
		},
		{
			name:   "RepoError",
			fields: fields{repo: NewInMemRepo()},
			args: args{
				ctx: context.Background(),
				id:  "test-route"},
			wantErr:   true,
			errType:   ErrorNotFound,
			cancelCtx: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{
				repo: tt.fields.repo,
			}

			c := tt.args.ctx
			cancel := func() {}
			if tt.cancelCtx {
				c, cancel = context.WithCancel(tt.args.ctx)
				cancel()
			}
			err := m.DeleteRoute(c, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteRoute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.wantErr {
				if !errors.Is(err, tt.errType) {
					t.Errorf("CreateRoute() returned error should be %s, but is %s", tt.errType, err)
				}
			}
		})
	}
}
