package transfer

import (
	"strings"
	"testing"

	"github.com/danlamanna/rivet/girder"
	"github.com/sirupsen/logrus"
)

func Test_collectResources(t *testing.T) {
	type args struct {
		ctx            *girder.Context
		localResources []string
	}
	tests := []struct {
		name string
		args args
		want []girder.Resource
	}{
		{
			name: "nonexistentfolder",
			args: args{
				ctx: &girder.Context{
					Logger: logrus.New(),
				},
				localResources: []string{"nonexistentfolder"},
			},
			want: []girder.Resource{},
		},
		{
			name: "using-dot",
			args: args{
				ctx: &girder.Context{
					Logger: logrus.New(),
				},
				localResources: []string{"../etc/testdata/single-file"},
			},
			want: []girder.Resource{
				{Path: "../etc/testdata/single-file", Type: "directory", Size: 4096},
				{Path: "../etc/testdata/single-file/.gitkeep", Type: "file", Size: 0},
			},
		},
		{
			name: "mixed testdata",
			args: args{
				ctx: &girder.Context{
					Logger: logrus.New(),
				},
				localResources: []string{"../etc/testdata/mixed"},
			},
			want: []girder.Resource{
				{Path: "../etc/testdata/mixed", Type: "directory", Size: 4096},
				{Path: "../etc/testdata/mixed/a", Type: "directory", Size: 4096},
				{Path: "../etc/testdata/mixed/a/b", Type: "directory", Size: 4096},
				{Path: "../etc/testdata/mixed/a/inner_file", Type: "file", Size: 4},
				{Path: "../etc/testdata/mixed/a_symlink", Type: "file", Size: 8},
				{Path: "../etc/testdata/mixed/root_file", Type: "file", Size: 0},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSlice := make([]girder.Resource, len(tt.want))
			gotChannel := collectResources(tt.args.ctx, tt.args.localResources...)
			i := 0
			for value := range gotChannel {
				// vscode debugger creates this file, https://github.com/Microsoft/vscode-go/issues/1301
				if !strings.Contains(value.Path, "debug.test") {
					gotSlice[i] = *value
					i++
				}
			}

			for i := 0; i < len(gotSlice) && i < len(tt.want); i++ {
				if gotSlice[i].Path != tt.want[i].Path ||
					gotSlice[i].Type != tt.want[i].Type ||
					gotSlice[i].Size != tt.want[i].Size {
					t.Errorf("collectResources(%d) = %v, want %v", i, gotSlice[i], tt.want[i])
				}
			}
		})
	}
}
