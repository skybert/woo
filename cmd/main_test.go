package main

import (
	"github.com/stretchr/testify/require"
	"testing"
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

func Test_expandTemplate(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		template []byte
		values   map[string]string
		want     string
	}{
		{
			name:     "expand one variable",
			template: []byte("<h1>{{TITLE}}</h1>"),
			values:   map[string]string{"TITLE": "This is my title"},
			want:     "<h1>This is my title</h1>",
		},
		{
			name:     "expand multiple variables",
			template: []byte("<a href=\"{{SITE_URL}}\">{{TITLE}}</a>"),
			values: map[string]string{
				"TITLE":    "My title",
				"SITE_URL": "https://example.com",
			},
			want: "<a href=\"https://example.com\">My title</a>",
		},
		{
			name:     "string with var name, but no variable syntax",
			template: []byte("No variables here called TITLE"),
			values:   map[string]string{"TITLE": "This is my title"},
			want:     "No variables here called TITLE",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expected := tt.want
			actual := expandTemplate(tt.template, tt.values)
			require.Equal(t, expected, actual, "failed "+tt.name)
		})
	}
}
