package configure

import (
	"context"
	"errors"
	"io/ioutil"
	"testing"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/fwiedmann/prox/domain/entity/route"
)

func Test_file_StartConfigure(t *testing.T) {
	t.Parallel()
	type fields struct {
		input        []route.Route
		routeManager route.Manager
		cancelTimout time.Duration
		fileType     string
	}
	tests := []struct {
		name        string
		fields      fields
		wantErr     bool
		wantErrType error
	}{
		{
			name: "ValidConfig",
			fields: fields{
				fileType: ".yaml",
				input: []route.Route{
					{
						NameID:   "test-1",
						Hostname: "docker.com",
					},
					{
						NameID:   "test-2",
						Hostname: "docker.com",
					},
				},
				routeManager: route.NewManager(route.NewInMemRepo(), route.CreateHTTPClientForRoute),
				cancelTimout: 1 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "InvalidConfigFileTyp",
			fields: fields{
				fileType: ".json",
				input: []route.Route{
					{
						NameID:   "test-1",
						Hostname: "docker.com",
					},
					{
						NameID:   "test-2",
						Hostname: "docker.com",
					},
				},
				routeManager: route.NewManager(route.NewInMemRepo(), route.CreateHTTPClientForRoute),
				cancelTimout: 1 * time.Second,
			},
			wantErr:     true,
			wantErrType: ErrorInvalidFileType,
		},
		{
			name: "InvalidConfigDuplicated",
			fields: fields{
				fileType: ".yaml",
				input: []route.Route{
					{
						NameID:   "test-1",
						Hostname: "docker.com",
					},
					{
						NameID:   "test-1",
						Hostname: "docker.com",
					},
				},
				routeManager: route.NewManager(route.NewInMemRepo(), route.CreateHTTPClientForRoute),
				cancelTimout: 1 * time.Second,
			},
			wantErr:     false,
			wantErrType: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			testFile, err := ioutil.TempFile(t.TempDir(), "*"+tt.fields.fileType)
			if err != nil {
				t.Error(err)
				return
			}

			body, err := yaml.Marshal(&tt.fields.input)
			if err != nil {
				t.Error(err)
				return
			}

			if err := ioutil.WriteFile(testFile.Name(), body, 0777); err != nil {
				t.Error(err)
				return
			}
			testFile.Close()

			f := &file{
				pathToFile:   testFile.Name(),
				routeManager: tt.fields.routeManager,
			}

			ctx, cancel := context.WithTimeout(context.Background(), tt.fields.cancelTimout)
			defer cancel()
			errChan := make(chan error, 2)
			go f.StartConfigure(ctx, errChan)

			select {
			case err := <-errChan:
				if (err != nil) && !tt.wantErr {
					t.Errorf("StartConfigure() send error: %s", err)
					return
				}
				if !errors.Is(err, tt.wantErrType) {
					t.Errorf("StartConfigure() send error %s, want %s", err, tt.wantErrType)
					return
				}
			case <-ctx.Done():
				if tt.wantErr {
					t.Errorf("StartConfigure() should have send an error")
					return
				}
			}

		})
	}
}
