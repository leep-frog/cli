package commands

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCompletors(t *testing.T) {
	for _, test := range []struct {
		name  string
		c     *Completor
		value *Value
		args  map[string]*Value
		flags map[string]*Value
		want  []string
	}{
		{
			name: "nil completor returns nil",
		},
		{
			name: "nil fetcher returns nil",
			c:    &Completor{},
		},
		{
			name: "non-distinct completor returns duplicates",
			c: &Completor{
				SuggestionFetcher: &ListFetcher{
					Options: []string{"first", "second", "third"},
				},
			},
			value: &Value{stringList: []string{"first", "second"}},
			want:  []string{"first", "second", "third"},
		},
		{
			name: "distinct completor returns duplicates",
			c: &Completor{
				Distinct: true,
				SuggestionFetcher: &ListFetcher{
					Options: []string{"first", "second", "third"},
				},
			},
			value: &Value{stringList: []string{"first", "second"}},
			want:  []string{"third"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := test.c.Complete(test.value, test.args, test.flags)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("completor.Complete(%v, %v, %v) returned diff (-want, +got):\n%s", test.value, test.args, test.flags, diff)
			}
		})
	}
}

func TestFetchers(t *testing.T) {
	for _, test := range []struct {
		name     string
		f        Fetcher
		value    *Value
		args     map[string]*Value
		flags    map[string]*Value
		getwdDir string
		getwdErr error
		want     []string
	}{
		{
			name: "noop fetcher returns nil",
			f:    &NoopFetcher{},
		},
		{
			name: "list fetcher returns nil",
			f:    &ListFetcher{},
		},
		{
			name: "list fetcher returns empty list",
			f: &ListFetcher{
				Options: []string{},
			},
			want: []string{},
		},
		{
			name: "list fetcher returns list",
			f: &ListFetcher{
				Options: []string{"first", "second", "third"},
			},
			want: []string{"first", "second", "third"},
		},
		// FileFetcher tests
		{
			name:     "file fetcher returns nil if failure fetching current directory",
			f:        &FileFetcher{},
			getwdErr: fmt.Errorf("failed to fetch directory"),
		},
		{
			name:     "file fetcher returns files in the current working directory",
			f:        &FileFetcher{},
			getwdDir: "testing",
			want: []string{
				"dir1",
				"dir2",
				"four.txt",
				"one.txt",
				"three.txt",
				"two.txt",
			},
		},
		{
			name:     "file fetcher returns nil if failure listing directory",
			f:        &FileFetcher{},
			getwdDir: "does-not-exist",
		},
		{
			name: "file fetcher returns files in the specified directory",
			f: &FileFetcher{
				Directory: "testing/dir1",
			},
			want: []string{
				"first.txt",
				"fourth.py",
				"second.py",
				"third.go",
			},
		},
		{
			name: "file fetcher returns files matching regex",
			f: &FileFetcher{
				Directory: "testing/dir1",
				Regexp:    regexp.MustCompile(".*.py$"),
			},
			want: []string{
				"fourth.py",
				"second.py",
			},
		},
		{
			name: "file fetcher ignores files",
			f: &FileFetcher{
				Directory:         "testing/dir2",
				IgnoreDirectories: true,
			},
			want: []string{
				"file1.txt",
				"file2.txt",
				"file3.txt",
			},
		},
		{
			name: "file fetcher ignores directories",
			f: &FileFetcher{
				Directory:   "testing/dir2",
				IgnoreFiles: true,
			},
			want: []string{
				"childC",
				"childD",
				"subA",
				"subB",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			oldWd := Getwd
			Getwd = func() (string, error) { return test.getwdDir, test.getwdErr }
			defer func() { Getwd = oldWd }()

			got := test.f.Fetch(test.value, test.args, test.flags)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("fetcher.Fetch(%v, %v, %v) returned diff (-want, +got):\n%s", test.value, test.args, test.flags, diff)
			}
		})
	}
}
