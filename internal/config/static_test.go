package config

import (
	"io/ioutil"
	"reflect"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestParseStaticFile(t *testing.T) {
	t.Parallel()
	type args struct {
		input        Static
		fileTypeName string
	}
	tests := []struct {
		name    string
		args    args
		want    Static
		wantErr bool
	}{
		{
			name: "Valid",
			args: args{
				input: Static{
					Ports: []Port{
						{Name: "test", Addr: 8080, TlSEnabled: true},
						{Name: "test2", Addr: 8081, TlSEnabled: true},
					},
					Cache: Cache{
						Enabled:                true,
						CacheMaxSizeInMegaByte: 0,
					},
				},
				fileTypeName: ".yaml",
			},
			want: Static{
				Ports: []Port{
					{Name: "test", Addr: 8080, TlSEnabled: true},
					{Name: "test2", Addr: 8081, TlSEnabled: true},
				},
				Cache: Cache{
					Enabled:                true,
					CacheMaxSizeInMegaByte: 0,
				},
				InfraPort: 9100,
			},
			wantErr: false,
		},
		{
			name: "InvalidDuplicated",
			args: args{
				input: Static{
					Ports: []Port{
						{Name: "test", Addr: 8080, TlSEnabled: true},
						{Name: "test", Addr: 8080, TlSEnabled: true},
					},
					Cache: Cache{
						Enabled:                true,
						CacheMaxSizeInMegaByte: 0,
					},
				},
				fileTypeName: ".yaml",
			},
			want:    Static{},
			wantErr: true,
		},
		{
			name: "InvalidDuplicated",
			args: args{
				input: Static{
					Ports: []Port{
						{Name: "test", Addr: 8080, TlSEnabled: true},
					},
					Cache: Cache{
						Enabled:                true,
						CacheMaxSizeInMegaByte: 0,
					},
				},
				fileTypeName: ".json",
			},
			want:    Static{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			body, err := yaml.Marshal(tt.args.input)
			if err != nil {
				t.Error(err)
				return
			}

			testFile, err := ioutil.TempFile(t.TempDir(), "*"+tt.args.fileTypeName)
			if err != nil {
				t.Error(err)
				return
			}

			if err := ioutil.WriteFile(testFile.Name(), body, 0777); err != nil {
				t.Error(err)
				return
			}

			got, err := ParseStaticFile(testFile.Name())
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseStaticFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseStaticFile() got = %v, want %v", got, tt.want)
			}
		})
	}
}
