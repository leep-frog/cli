package commands

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCompletors(t *testing.T) {
	for _, test := range []struct {
		name string
		c    *Completor
		args []string
		want []string
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
			args: []string{"first", "second", ""},
			want: []string{"first", "second", "third"},
		},
		{
			name: "distinct completor does not return duplicates",
			c: &Completor{
				Distinct: true,
				SuggestionFetcher: &ListFetcher{
					Options: []string{"first", "second", "third"},
				},
			},
			args: []string{"first", "second", ""},
			want: []string{"third"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			cmd := &TerminusCommand{
				Args: []Arg{
					StringListArg("test", 2, 5, test.c),
				},
			}
			got := Autocomplete(cmd, test.args, 0)
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("Autocomplete(%v, %v) returned diff (-want, +got):\n%s", cmd, test.args, diff)
			}
		})
	}
}

func TestFetchers(t *testing.T) {
	for _, test := range []struct {
		name     string
		f        Fetcher
		args     []string
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
			oldWd := getwd
			getwd = func() (string, error) { return test.getwdDir, test.getwdErr }
			defer func() { getwd = oldWd }()

			completor := &Completor{
				SuggestionFetcher: test.f,
			}
			cmd := &TerminusCommand{
				Args: []Arg{
					StringListArg("test", 2, 5, completor),
				},
			}
			got := Autocomplete(cmd, test.args, 0)
			if len(got) == 0 {
				got = nil
			}
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("Autocomplete(%v, %v) returned diff (-want, +got):\n%s", cmd, test.args, diff)
			}
		})
	}
}
