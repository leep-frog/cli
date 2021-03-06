package commands

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestAliaserAutocomplete(t *testing.T) {
	for _, test := range []struct {
		name string
		ac   *AliasCommand
		args []string
		want []string
	}{
		{
			name: "suggests subcommands",
			ac:   &AliasCommand{},
			want: []string{"a", "d", "g", "l", "s"},
		},
		// DeleteAlias tests.
		{
			name: "DeleteAlias suggests aliases",
			ac: &AliasCommand{
				Aliases: map[string]*Value{
					"aliasOne":   boolVal(true),
					"aliasTwo":   boolVal(true),
					"aliasThree": boolVal(true),
					"aliasFour":  boolVal(true),
				},
			},
			args: []string{"d", ""},
			want: []string{
				"aliasFour",
				"aliasOne",
				"aliasThree",
				"aliasTwo",
			},
		},
		{
			name: "DeleteAlias suggests unique aliases",
			ac: &AliasCommand{
				Aliases: map[string]*Value{
					"aliasOne":   boolVal(true),
					"aliasTwo":   boolVal(true),
					"aliasThree": boolVal(true),
					"aliasFour":  boolVal(true),
				},
			},
			args: []string{"d", "aliasFour", "missing", "aliasTwo", "ali"},
			want: []string{
				"aliasOne",
				"aliasThree",
			},
		},
		// GetAlias tests.
		{
			name: "GetAlias suggests aliases",
			ac: &AliasCommand{
				Aliases: map[string]*Value{
					"aliasOne":   boolVal(true),
					"aliasTwo":   boolVal(true),
					"aliasThree": boolVal(true),
					"aliasFour":  boolVal(true),
				},
			},
			args: []string{"g", ""},
			want: []string{
				"aliasFour",
				"aliasOne",
				"aliasThree",
				"aliasTwo",
			},
		},
		{
			name: "GetAlias completes alias",
			ac: &AliasCommand{
				Aliases: map[string]*Value{
					"aliasOne":   boolVal(true),
					"aliasTwo":   boolVal(true),
					"aliasThree": boolVal(true),
					"aliasFour":  boolVal(true),
				},
			},
			args: []string{"g", "aliasF"},
			want: []string{
				"aliasFour",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if test.ac.Aliaser == nil {
				test.ac.Aliaser = &testAliaser{}
			}
			suggestions := Autocomplete(test.ac.Command(), test.args, -1)
			if diff := cmp.Diff(test.want, suggestions); diff != "" {
				t.Errorf("Complete(%v) produced diff (-want, +got):\n%s", test.args, diff)
			}
		})
	}
}

type testAliaser struct {
	arg       Arg
	validate  func(cos CommandOS, alias string, value *Value, args, flags map[string]*Value) bool
	transform func(cos CommandOS, alias string, value *Value, args, flags map[string]*Value) (*Value, bool)
}

func (ta *testAliaser) Validate(cos CommandOS, alias string, value *Value, args, flags map[string]*Value) bool {
	if ta.validate == nil {
		return true
	}
	return ta.validate(cos, alias, value, args, flags)
}

func (ta *testAliaser) Transform(cos CommandOS, alias string, value *Value, args, flags map[string]*Value) (*Value, bool) {
	if ta.transform == nil {
		return value, true
	}
	return ta.transform(cos, alias, value, args, flags)
}

func (ta *testAliaser) Arg() Arg {
	return ta.arg
}

func TestAliasCommandExecution(t *testing.T) {
	for _, test := range []struct {
		name       string
		ac         *AliasCommand
		args       []string
		want       *AliasCommand
		wantOK     bool
		wantResp   *ExecutorResponse
		wantStdout []string
		wantStderr []string
		osStatErr  error
		osStatInfo os.FileInfo
		abs        string
		absErr     error
	}{
		{
			name: "subcommand argument required",
			ac: &AliasCommand{
				Aliaser: &testAliaser{
					arg: StringArg("str", true, nil),
				},
			},
			wantStderr: []string{
				"more args required",
			},
		},
		// AddAlias tests.
		{
			name: "AddAlias requires alias arg",
			ac: &AliasCommand{
				Aliaser: &testAliaser{
					arg: StringArg("str", true, nil),
				},
			},
			args: []string{"a"},
			wantStderr: []string{
				`no argument provided for "ALIAS"`,
			},
		},
		{
			name: "AddAlias requires alias value arg",
			ac: &AliasCommand{
				Aliaser: &testAliaser{
					arg: StringArg("str", true, nil),
				},
			},
			args: []string{"a", "salt"},
			wantStderr: []string{
				`no argument provided for "str"`,
			},
		},
		{
			name: "AddAlias fails if alias already exists",
			ac: &AliasCommand{
				Aliaser: &testAliaser{
					arg: StringArg("str", true, nil),
				},
				Aliases: map[string]*Value{
					"salt": stringVal("NaCl"),
				},
			},
			args: []string{"a", "salt", "sodiumChloride"},
			wantStderr: []string{
				"alias already defined: (salt: NaCl)",
			},
		},
		{
			name: "AddAlias adds an alias to an empty map",
			ac: &AliasCommand{
				Aliaser: &testAliaser{
					arg: StringArg("str", true, nil),
				},
			},
			args:   []string{"a", "salt", "NaCl"},
			wantOK: true,
			want: &AliasCommand{
				Aliases: map[string]*Value{
					"salt": stringVal("NaCl"),
				},
			},
		},
		{
			name: "AddAlias adds an alias to an existing map",
			ac: &AliasCommand{
				Aliaser: &testAliaser{
					arg: StringArg("str", true, nil),
				},
				Aliases: map[string]*Value{
					"breakfast": stringList("green", "eggs", "and", "ham"),
				},
			},
			args:   []string{"a", "salt", "NaCl"},
			wantOK: true,
			want: &AliasCommand{
				Aliases: map[string]*Value{
					"breakfast": stringList("green", "eggs", "and", "ham"),
					"salt":      stringVal("NaCl"),
				},
			},
		},
		{
			name: "AddAlias fails if verifier fails",
			ac: &AliasCommand{
				Aliaser: &testAliaser{
					arg: StringArg("str", true, nil),
					validate: func(cos CommandOS, alias string, value *Value, args, flags map[string]*Value) bool {
						cos.Stderr("bad news bears")
						return false
					},
				},
			},
			args: []string{"a", "salt", "sodiumChloride"},
			wantStderr: []string{
				"bad news bears",
			},
		},
		{
			name: "AddAlias works if the verifier passes",
			ac: &AliasCommand{
				Aliaser: &testAliaser{
					arg: StringArg("str", true, nil),
					validate: func(cos CommandOS, alias string, value *Value, args, flags map[string]*Value) bool {
						cos.Stdout("good news tigers")
						return true
					},
				},
			},
			args:   []string{"a", "salt", "NaCl"},
			wantOK: true,
			wantStdout: []string{
				"good news tigers",
			},
			want: &AliasCommand{
				Aliases: map[string]*Value{
					"salt": stringVal("NaCl"),
				},
			},
		},
		{
			name: "AddAlias fails if the transformer fails",
			ac: &AliasCommand{
				Aliaser: &testAliaser{
					arg: StringArg("str", true, nil),
					transform: func(cos CommandOS, alias string, value *Value, args, flags map[string]*Value) (*Value, bool) {
						cos.Stderr("bad news lions")
						return nil, false
					},
				},
			},
			args: []string{"a", "salt", "NaCl"},
			wantStderr: []string{
				"bad news lions",
			},
		},
		{
			name: "AddAlias transforms the value",
			ac: &AliasCommand{
				Aliaser: &testAliaser{
					arg: StringArg("str", true, nil),
					transform: func(cos CommandOS, alias string, value *Value, args, flags map[string]*Value) (*Value, bool) {
						return stringList("Na", "Cl"), true
					},
				},
			},
			args:   []string{"a", "salt", "NaCl"},
			wantOK: true,
			want: &AliasCommand{
				Aliases: map[string]*Value{
					"salt": stringList("Na", "Cl"),
				},
			},
		},
		// DeleteAlias tests.
		{
			name: "DeleteAlias requires at least one arg",
			ac:   &AliasCommand{},
			args: []string{"d"},
			wantStderr: []string{
				`no argument provided for "ALIAS"`,
			},
		},
		{
			name: "DeleteAlias handles nonexistent aliases",
			ac: &AliasCommand{
				Aliases: map[string]*Value{
					"salt": stringList("Na", "Cl"),
				},
			},
			args:   []string{"d", "pepper"},
			wantOK: true,
			wantStderr: []string{
				`alias "pepper" does not exist`,
			},
		},
		{
			name: "DeleteAlias deletes alias",
			ac: &AliasCommand{
				Aliases: map[string]*Value{
					"salt":   stringList("Na", "Cl"),
					"pepper": stringVal("sneezy"),
				},
			},
			args:   []string{"d", "pepper"},
			wantOK: true,
			want: &AliasCommand{
				Aliases: map[string]*Value{
					"salt": stringList("Na", "Cl"),
				},
			},
		},
		{
			name: "DeleteAlias handles several args",
			ac: &AliasCommand{
				Aliases: map[string]*Value{
					"salt":   stringList("Na", "Cl"),
					"pepper": stringVal("sneezy"),
				},
			},
			args:   []string{"d", "garlic", "pepper", "other"},
			wantOK: true,
			want: &AliasCommand{
				Aliases: map[string]*Value{
					"salt": stringList("Na", "Cl"),
				},
			},
			wantStderr: []string{
				`alias "garlic" does not exist`,
				`alias "other" does not exist`,
			},
		},
		// GetAlias tests.
		{
			name: "GetAlias requires alias arg",
			ac:   &AliasCommand{},
			args: []string{"g"},
			wantStderr: []string{
				`no argument provided for "ALIAS"`,
			},
		},
		{
			name: "GetAlias fails if alias does not exist",
			ac:   &AliasCommand{},
			args: []string{"g", "pepper"},
			wantStderr: []string{
				`Alias "pepper" does not exist`,
			},
		},
		{
			name: "GetAlias gets an alias",
			ac: &AliasCommand{
				Aliases: map[string]*Value{
					"salt": stringList("Na", "Cl"),
				},
			},
			wantOK: true,
			args:   []string{"g", "salt"},
			wantStdout: []string{
				"salt: Na, Cl",
			},
		},
		// ListAliases tests.
		{
			name: "ListAliases lists the aliases",
			ac: &AliasCommand{
				Aliases: map[string]*Value{
					"salt":    stringList("Na", "Cl"),
					"pepper":  stringVal("sneezy"),
					"oregano": boolVal(false),
					"garlic":  intVal(2468),
					"curry":   floatVal(-13.57),
				},
			},
			args:   []string{"l"},
			wantOK: true,
			wantStdout: []string{
				"curry: -13.57",
				"garlic: 2468",
				"oregano: false",
				"pepper: sneezy",
				"salt: Na, Cl",
			},
		},
		// SearchAlias tests.
		{
			name: "SearchAlias requires a regex",
			ac:   &AliasCommand{},
			args: []string{"s"},
			wantStderr: []string{
				`no argument provided for "REGEXP"`,
			},
		},
		{
			name: "SearchAlias requires a valid regex",
			ac:   &AliasCommand{},
			args: []string{"s", ":)"},
			wantStderr: []string{
				"Invalid regexp: error parsing regexp: unexpected ): `:)`",
			},
		},
		{
			name: "SearchAlias works",
			ac: &AliasCommand{
				Aliases: map[string]*Value{
					"salt":    stringList("Na", "Cl"),
					"pepper":  stringVal("sneezy"),
					"oregano": boolVal(false),
					"garlic":  intVal(2468),
					"curry":   floatVal(-13.57),
				},
			},
			args:   []string{"s", "^......:"},
			wantOK: true,
			wantStdout: []string{
				"garlic: 2468",
				"pepper: sneezy",
			},
		},
		// FileAliaser tests (only need to test AddAlias).
		{
			name: "FileAliaser fails if stat error in validate",
			ac: &AliasCommand{
				Aliaser: NewFileAliaser(),
			},
			args:      []string{"a", "shortcut", "the-low-road"},
			osStatErr: fmt.Errorf("oops"),
			wantStderr: []string{
				"file does not exist: oops",
			},
		},
		{
			name: "FileAliaser fails if filepathAbs error in transform",
			ac: &AliasCommand{
				Aliaser: NewFileAliaser(),
			},
			args:       []string{"a", "shortcut", "the-low-road"},
			osStatInfo: file(),
			absErr:     fmt.Errorf("absolutely not"),
			wantStderr: []string{
				`failed to get absolute file path for file "the-low-road": absolutely not`,
			},
		},
		{
			name: "FileAliaser adds file alias",
			ac: &AliasCommand{
				Aliaser: NewFileAliaser(),
			},
			args:       []string{"a", "shortcut", "the-low-road"},
			osStatInfo: file(),
			abs:        "scotland/the-low-road",
			wantOK:     true,
			want: &AliasCommand{
				Aliases: map[string]*Value{
					"shortcut": stringVal("scotland/the-low-road"),
				},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			oldStat := osStat
			osStat = func(_ string) (os.FileInfo, error) { return test.osStatInfo, test.osStatErr }
			defer func() { osStat = oldStat }()

			oldAbs := filepathAbs
			filepathAbs = func(_ string) (string, error) { return test.abs, test.absErr }
			defer func() { filepathAbs = oldAbs }()

			if test.ac.Aliaser == nil {
				test.ac.Aliaser = &testAliaser{}
			}
			tcos := &TestCommandOS{}
			got, ok := Execute(tcos, test.ac.Command(), test.args, nil)
			if ok != test.wantOK {
				t.Fatalf("commands.Execute(%v) returned %v for ok; want %v", test.args, ok, test.wantOK)
			}
			if diff := cmp.Diff(test.wantResp, got); diff != "" {
				t.Fatalf("Execute(%v) produced response diff (-want, +got):\n%s", test.args, diff)
			}

			if diff := cmp.Diff(test.wantStdout, tcos.GetStdout()); diff != "" {
				t.Errorf("command.Execute(%v) produced stdout diff (-want, +got):\n%s", test.args, diff)
			}
			if diff := cmp.Diff(test.wantStderr, tcos.GetStderr()); diff != "" {
				t.Errorf("command.Execute(%v) produced stderr diff (-want, +got):\n%s", test.args, diff)
			}

			// Assume wantChanged if test.want is set
			wantChanged := test.want != nil
			changed := test.ac != nil && test.ac.Changed()
			if changed != wantChanged {
				t.Fatalf("Execute(%v) marked Changed as %v; want %v", test.args, changed, wantChanged)
			}

			// Only check diff if we are expecting a change.
			if wantChanged {
				opts := []cmp.Option{
					cmpopts.IgnoreUnexported(AliasCommand{}, genericArgs{}, Value{}, StringList{}, testAliaser{}),
					cmpopts.IgnoreFields(AliasCommand{}, "Aliaser"),
				}
				if diff := cmp.Diff(test.want, test.ac, opts...); diff != "" {
					t.Fatalf("Execute(%v) produced emacs diff (-want, +got):\n%s", test.args, diff)
				}
			}
		})
	}
}

func file() os.FileInfo {
	return &fakeFileInfo{mode: 0}
}

func directory() os.FileInfo {
	return &fakeFileInfo{mode: os.ModeDir}
}

type fakeFileInfo struct{ mode os.FileMode }

func (fi fakeFileInfo) Name() string       { return "" }
func (fi fakeFileInfo) Size() int64        { return 0 }
func (fi fakeFileInfo) Mode() os.FileMode  { return fi.mode }
func (fi fakeFileInfo) ModTime() time.Time { return time.Now() }
func (fi fakeFileInfo) IsDir() bool        { return fi.Mode().IsDir() }
func (fi fakeFileInfo) Sys() interface{}   { return nil }
