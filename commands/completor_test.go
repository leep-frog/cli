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
		name          string
		f             Fetcher
		distinct      bool
		args          []string
		absErr        error
		stringArg     bool
		commandBranch bool
		want          []string
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
		// TODO: automatically create the empty directory
		// at beginning of this test (since not tracked by git).
		{
			name: "file fetcher handles empty directory",
			f:    &FileFetcher{},
			args: []string{"testing/empty/"},
		},
		{
			name: "file fetcher returns files in the current working directory",
			f:    &FileFetcher{},
			want: []string{
				"aliaser.go",
				"aliaser_test.go",
				"arg_options.go",
				"arg_types.go",
				"commands.go",
				"commands_test.go",
				"completor_test.go",
				"completors.go",
				"flag_types.go",
				"new_arg_types.go",
				"README.md",
				"testing/",
				"value.proto",
				"value/",
				"value_test.go",
				"values.go",
				" ",
			},
		},
		{
			name: "file fetcher works with string list arg",
			f:    &FileFetcher{},
			args: []string{"ar"},
			want: []string{
				"arg_",
				"arg__",
			},
		},
		{
			name:      "file fetcher works with string arg",
			f:         &FileFetcher{},
			args:      []string{"ar"},
			stringArg: true,
			want: []string{
				"arg_",
				"arg__",
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
				Directory: "testing",
			},
			want: []string{
				".surprise",
				"cases/",
				"dir1/",
				"dir2/",
				"dir3/",
				"dir4/",
				"empty/",
				"four.txt",
				"METADATA",
				"metadata_/",
				"moreCases/",
				"one.txt",
				"three.txt",
				"two.txt",
				" ",
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
				" ",
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
				" ",
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
				" ",
			},
		},
		{
			name: "file fetcher ignores directories",
			f: &FileFetcher{
				Directory:         "testing/dir2",
				IgnoreDirectories: true,
			},
			want: []string{
				"file",
				"file_",
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
				" ",
			},
		},
		{
			name: "file fetcher completes to directory",
			f:    &FileFetcher{},
			args: []string{"testing/dir1"},
			want: []string{
				"testing/dir1/",
				"testing/dir1/_",
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
				"dir1/_",
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
				" ",
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
				" ",
			},
		},
		{
			name: "file fetcher only shows basenames when multiple options with different next letter",
			f:    &FileFetcher{},
			args: []string{"testing/dir"},
			want: []string{
				"dir1/",
				"dir2/",
				"dir3/",
				"dir4/",
				" ",
			},
		},
		{
			name: "file fetcher shows full names when multiple options with same next letter",
			f:    &FileFetcher{},
			args: []string{"testing/d"},
			want: []string{
				"testing/dir",
				"testing/dir_",
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
				" ",
			},
		},
		{
			name: "file fetcher handles directories with spaces",
			f:    &FileFetcher{},
			args: []string{`testing/dir4/folder\`, `wit`},
			want: []string{
				`testing/dir4/folder\ with\ spaces/`,
				`testing/dir4/folder\ with\ spaces/_`,
			},
		},
		{
			name: "file fetcher handles directories with spaces when same argument",
			f:    &FileFetcher{},
			args: []string{`testing/dir4/folder\ wit`},
			want: []string{
				`testing/dir4/folder\ with\ spaces/`,
				`testing/dir4/folder\ with\ spaces/_`,
			},
		},
		{
			name: "file fetcher can dive into folder with spaces",
			f:    &FileFetcher{},
			args: []string{`testing/dir4/folder\`, `with\`, `spaces/`},
			want: []string{
				"goodbye.go",
				"hello.txt",
				" ",
			},
		},
		{
			name: "file fetcher can dive into folder with spaces when combined args",
			f:    &FileFetcher{},
			args: []string{`testing/dir4/folder\ with\ spaces/`},
			want: []string{
				"goodbye.go",
				"hello.txt",
				" ",
			},
		},
		{
			name: "autocomplete fills in letters that are the same for all options",
			f:    &FileFetcher{},
			args: []string{`testing/dir4/fo`},
			want: []string{
				"testing/dir4/folder",
				"testing/dir4/folder_",
			},
		},
		{
			name:          "file fetcher doesn't get filtered out when part of a CommandBranch",
			f:             &FileFetcher{},
			commandBranch: true,
			args:          []string{"testing/dir"},
			want: []string{
				"dir1/",
				"dir2/",
				"dir3/",
				"dir4/",
				" ",
			},
		},
		{
			name:          "file fetcher handles multiple options in directory",
			f:             &FileFetcher{},
			commandBranch: true,
			args:          []string{"testing/dir1/f"},
			want: []string{
				"first.txt",
				"fourth.py",
				" ",
			},
		},
		{
			name:          "case insensitive gets letters autofilled",
			f:             &FileFetcher{},
			commandBranch: true,
			args:          []string{"testing/dI"},
			want: []string{
				"testing/dir",
				"testing/dir_",
			},
		},
		{
			name:          "case insensitive recommends all without complete",
			f:             &FileFetcher{},
			commandBranch: true,
			args:          []string{"testing/DiR"},
			want: []string{
				"dir1/",
				"dir2/",
				"dir3/",
				"dir4/",
				" ",
			},
		},
		{
			name:          "file fetcher ignores case",
			f:             &FileFetcher{},
			commandBranch: true,
			args:          []string{"testing/cases/abc"},
			want: []string{
				"testing/cases/abcde",
				"testing/cases/abcde_",
			},
		},
		{
			name:          "file fetcher sorting ignores cases when no file",
			f:             &FileFetcher{},
			commandBranch: true,
			args:          []string{"testing/moreCases/"},
			want: []string{
				"testing/moreCases/QW_",
				"testing/moreCases/QW__",
			},
		},
		{
			name:          "file fetcher sorting ignores cases when autofilling",
			f:             &FileFetcher{},
			commandBranch: true,
			args:          []string{"testing/moreCases/q"},
			want: []string{
				"testing/moreCases/qW_",
				"testing/moreCases/qW__",
			},
		},
		{
			name:          "file fetcher sorting ignores cases when not autofilling",
			f:             &FileFetcher{},
			commandBranch: true,
			args:          []string{"testing/moreCases/qW_t"},
			want: []string{
				"qW_three.txt",
				"qw_TRES.txt",
				"Qw_two.txt",
				" ",
			},
		},
		{
			name: "file fetcher completes to case matched completion",
			f:    &FileFetcher{},
			args: []string{"testing/meta"},
			want: []string{
				"testing/metadata",
				"testing/metadata_",
			},
		},
		{
			name: "file fetcher completes to case matched completion",
			f:    &FileFetcher{},
			args: []string{"testing/ME"},
			want: []string{
				"testing/METADATA",
				"testing/METADATA_",
			},
		},
		{
			name: "file fetcher completes to something when no cases match",
			f:    &FileFetcher{},
			args: []string{"testing/MeTa"},
			want: []string{
				"testing/METADATA",
				"testing/METADATA_",
			},
		},
		{
			name: "file fetcher completes to case matched completion in current directory",
			f: &FileFetcher{
				Directory: "testing",
			},
			args: []string{"meta"},
			want: []string{
				"metadata",
				"metadata_",
			},
		},
		{
			name: "file fetcher completes to case matched completion in current directory",
			f: &FileFetcher{
				Directory: "testing",
			},
			args: []string{"MET"},
			want: []string{
				"METADATA",
				"METADATA_",
			},
		},
		{
			name: "file fetcher completes to something when no cases match in current directory",
			f: &FileFetcher{
				Directory: "testing",
			},
			args: []string{"meTA"},
			want: []string{
				"METADATA",
				"METADATA_",
			},
		},
		{
			name: "file fetcher doesn't complete when matches a prefix",
			f:    &FileFetcher{},
			args: []string{"testing/METADATA"},
			want: []string{
				"METADATA",
				"metadata_/",
				" ",
			},
		},
		{
			name: "file fetcher doesn't complete when matches a prefix file",
			f:    &FileFetcher{},
			args: []string{"testing/metadata_/m"},
			want: []string{
				"m1",
				"m2",
				" ",
			},
		},
		{
			name:     "file fetcher returns complete match if distinct",
			f:        &FileFetcher{},
			distinct: true,
			args:     []string{"testing/metadata_/m1"},
			want: []string{
				"testing/metadata_/m1",
			},
		},
		// Distinct file fetchers.
		{
			name: "file fetcher returns repeats if not distinct",
			f:    &FileFetcher{},
			args: []string{"testing/three.txt", "testing/t"},
			want: []string{"three.txt", "two.txt", " "},
		},
		{
			name: "file fetcher returns distinct",
			f: &FileFetcher{
				Distinct: true,
			},
			args: []string{"testing/three.txt", "testing/t"},
			want: []string{"testing/two.txt"},
		},
		{
			name: "file fetcher handles non with distinct",
			f: &FileFetcher{
				Distinct: true,
			},
			args: []string{"testing/three.txt", "testing/two.txt", "testing/t"},
		},
		{
			name: "file fetcher first level distinct partially completes",
			f: &FileFetcher{
				Distinct: true,
			},
			args: []string{"c"},
			want: []string{"com", "com_"},
		},
		{
			name: "file fetcher first level distinct returns all options",
			f: &FileFetcher{
				Distinct: true,
			},
			args: []string{"com"},
			want: []string{
				"commands.go",
				"commands_test.go",
				"completor_test.go",
				"completors.go",
				" ",
			},
		},
		{
			name: "file fetcher first level distinct completes partial",
			f: &FileFetcher{
				Distinct: true,
			},
			args: []string{"commands.go", "c"},
			want: []string{
				"com",
				"com_",
			},
		},
		{
			name: "file fetcher first level distinct suggests remaining",
			f: &FileFetcher{
				Distinct: true,
			},
			args: []string{"commands.go", "com"},
			want: []string{
				"commands_test.go",
				"completor_test.go",
				"completors.go",
				" ",
			},
		},
		{
			name: "file fetcher first level distinct completes partial",
			f: &FileFetcher{
				Distinct: true,
			},
			args: []string{"commands.go", "commands_test.go", "c"},
			want: []string{
				"completor",
				"completor_",
			},
		},
		{
			name: "file fetcher first level distinct suggests remaining subset",
			f: &FileFetcher{
				Distinct: true,
			},
			args: []string{"commands.go", "commands_test.go", "completor"},
			want: []string{
				"completor_test.go",
				"completors.go",
				" ",
			},
		},
		{
			name: "file fetcher first level distinct autofills remaining",
			f: &FileFetcher{
				Distinct: true,
			},
			args: []string{"commands.go", "commands_test.go", "completor_"},
			want: []string{
				"completor_test.go",
			},
		},
		/* Useful for commenting out tests */
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
				Distinct:          test.distinct,
			}

			arg := StringListArg("test", 2, 5, completor)
			if test.stringArg {
				arg = StringArg("test", true, completor)
			}
			var cmd Command
			tc := &TerminusCommand{Args: []Arg{arg}}
			if test.commandBranch {
				cmd = &CommandBranch{
					TerminusCommand: tc,
				}
			} else {
				cmd = tc
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
