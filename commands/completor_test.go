package commands

import (
	"fmt"
	"path/filepath"
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
		name   string
		f      Fetcher
		args   []string
		absErr error
		want   []string
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
			name:   "file fetcher returns nil if failure fetching current directory",
			f:      &FileFetcher{},
			absErr: fmt.Errorf("failed to fetch directory"),
		},
		{
			name: "file fetcher returns files in the current working directory",
			f:    &FileFetcher{},
			want: []string{
				"arg_options.go",
				"arg_types.go",
				"commands.go",
				"commands_test.go",
				"completor_test.go",
				"completors.go",
				"flag_types.go",
				"testing/",
				"value_test.go",
				"values.go",
			},
		},
		{
			name: "file fetcher returns nil if failure listing directory",
			f: &FileFetcher{
				Directory: "does/not/exist",
			},
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
			name: "file fetcher requires prefix",
			f: &FileFetcher{
				Directory: "testing/dir3",
			},
			args: []string{"th"},
			want: []string{
				"that/",
				"this.txt",
			},
		},
		{
			name: "file fetcher ignores directories",
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
			name: "file fetcher ignores files",
			f: &FileFetcher{
				Directory:   "testing/dir2",
				IgnoreFiles: true,
			},
			want: []string{
				"childC/",
				"childD/",
				"subA/",
				"subB/",
			},
		},
		{
			name: "file fetcher completes to directory",
			f:    &FileFetcher{},
			args: []string{"testing/dir1"},
			want: []string{
				"testing/dir1/",
				"testing/dir1//",
			},
		},
		{
			name: "file fetcher completes to directory when starting dir specified",
			f: &FileFetcher{
				Directory: "testing",
			},
			args: []string{"dir1"},
			want: []string{
				"dir1/",
				"dir1//",
			},
		},
		{
			name: "file fetcher shows contents of directory when ending with a separator",
			f:    &FileFetcher{},
			args: []string{"testing/dir1/"},
			want: []string{
				"first.txt",
				"fourth.py",
				"second.py",
				"third.go",
			},
		},
		{
			name: "file fetcher completes to directory when ending with a separator and when starting dir specified",
			f: &FileFetcher{
				Directory: "testing",
			},
			args: []string{"dir1/"},
			want: []string{
				"first.txt",
				"fourth.py",
				"second.py",
				"third.go",
			},
		},
		{
			name: "file fetcher only shows basenames when multiple options",
			f:    &FileFetcher{},
			args: []string{"testing/di"},
			want: []string{
				"dir1/",
				"dir2/",
				"dir3/",
			},
		},
		{
			name: "file fetcher only shows basenames when multiple options and starting dir",
			f: &FileFetcher{
				Directory: "testing/dir1",
			},
			args: []string{"f"},
			want: []string{
				"first.txt",
				"fourth.py",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			oldAbs := filepathAbs
			filepathAbs = func(rel string) (string, error) {
				if test.absErr != nil {
					return "", test.absErr
				}
				return filepath.Abs(rel)
			}
			defer func() { filepathAbs = oldAbs }()

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
