package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_getTargetDir(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		fn   string
		want string
	}{
		{
			name: "get target dir, happy path",
			fn:   "/usr/src/web/linux/emacs-go.md",
			want: "/tmp/output/linux",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getTargetDir(tt.fn)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestWoo_getTargetFile(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		dir     string
		fn      string
		want    string
		wantErr bool
	}{
		{
			name:    "get target file, happy path",
			dir:     "/var/www/linux",
			fn:      "/usr/src/web/linux/emacs.md",
			want:    "/var/www/linux/emacs.html",
			wantErr: false,
		},
		{
			name:    "get target file, not an md",
			dir:     "/var/www",
			fn:      "/usr/src/web/linux/emacs.txt",
			want:    "/var/www/linux/emacs.html",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWoo()
			got, gotErr := w.getTargetFile(tt.dir, tt.fn)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("getTargetFile() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("getTargetFile() succeeded unexpectedly")
			}
			require.Equal(t, tt.want, got)
		})
	}
}
