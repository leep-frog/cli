package commands

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestAliaserAutocomplete(t *testing.T) {
	for _, test := range []struct {
		name string
		a    *Aliaser
		args []string
		want []string
	}{
		{
			name: "suggests subcommands",
			a:    &Aliaser{},
			want: []string{"a", "d", "g", "l", "s"},
		},
		// DeleteAlias tests.
		{
			name: "DeleteAlias suggests aliases",
			a: &Aliaser{
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
			a: &Aliaser{
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
			a: &Aliaser{
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
			a: &Aliaser{
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
			suggestions := Autocomplete(test.a.Command(), test.args, -1)
			if diff := cmp.Diff(test.want, suggestions); diff != "" {
				t.Errorf("Complete(%v) produced diff (-want, +got):\n%s", test.args, diff)
			}
		})
	}
}

func TestAliaserExecution(t *testing.T) {
	for _, test := range []struct {
		name          string
		a             *Aliaser
		args          []string
		limitOverride int
		want          *Aliaser
		wantOK        bool
		wantResp      *ExecutorResponse
		wantStdout    []string
		wantStderr    []string
	}{
		{
			name: "subcommand argument required",
			a: &Aliaser{
				Arg: StringArg("str", true, nil),
			},
			wantStderr: []string{
				"more args required",
			},
		},
		// AddAlias tests.
		{
			name: "AddAlias requires alias arg",
			a: &Aliaser{
				Arg: StringArg("str", true, nil),
			},
			args: []string{"a"},
			wantStderr: []string{
				`no argument provided for "ALIAS"`,
			},
		},
		{
			name: "AddAlias requires alias value arg",
			a: &Aliaser{
				Arg: StringArg("str", true, nil),
			},
			args: []string{"a", "salt"},
			wantStderr: []string{
				`no argument provided for "str"`,
			},
		},
		{
			name: "AddAlias fails if alias already exists",
			a: &Aliaser{
				Arg: StringArg("str", true, nil),
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
			a: &Aliaser{
				Arg: StringArg("str", true, nil),
			},
			args:   []string{"a", "salt", "NaCl"},
			wantOK: true,
			want: &Aliaser{
				Aliases: map[string]*Value{
					"salt": stringVal("NaCl"),
				},
			},
		},
		{
			name: "AddAlias adds an alias to an existing map",
			a: &Aliaser{
				Arg: StringArg("str", true, nil),
				Aliases: map[string]*Value{
					"breakfast": stringList("green", "eggs", "and", "ham"),
				},
			},
			args:   []string{"a", "salt", "NaCl"},
			wantOK: true,
			want: &Aliaser{
				Aliases: map[string]*Value{
					"breakfast": stringList("green", "eggs", "and", "ham"),
					"salt":      stringVal("NaCl"),
				},
			},
		},
		{
			name: "AddAlias fails if verifier fails",
			a: &Aliaser{
				Arg: StringArg("str", true, nil),
				Verifier: func(cos CommandOS, alias string, value *Value, args, flags map[string]*Value) bool {
					cos.Stderr("bad news bears")
					return false
				},
			},
			args: []string{"a", "salt", "sodiumChloride"},
			wantStderr: []string{
				"bad news bears",
			},
		},
		{
			name: "AddAlias works if the verifier passes",
			a: &Aliaser{
				Arg: StringArg("str", true, nil),
				Verifier: func(cos CommandOS, alias string, value *Value, args, flags map[string]*Value) bool {
					cos.Stdout("good news tigers")
					return true
				},
			},
			args:   []string{"a", "salt", "NaCl"},
			wantOK: true,
			wantStdout: []string{
				"good news tigers",
			},
			want: &Aliaser{
				Aliases: map[string]*Value{
					"salt": stringVal("NaCl"),
				},
			},
		},
		{
			name: "AddAlias fails if the transformer fails",
			a: &Aliaser{
				Arg: StringArg("str", true, nil),
				Transformer: func(cos CommandOS, alias string, value *Value, args, flags map[string]*Value) (*Value, bool) {
					cos.Stderr("bad news lions")
					return nil, false
				},
			},
			args: []string{"a", "salt", "NaCl"},
			wantStderr: []string{
				"bad news lions",
			},
		},
		{
			name: "AddAlias transforms the value",
			a: &Aliaser{
				Arg: StringArg("str", true, nil),
				Transformer: func(cos CommandOS, alias string, value *Value, args, flags map[string]*Value) (*Value, bool) {
					return stringList("Na", "Cl"), true
				},
			},
			args:   []string{"a", "salt", "NaCl"},
			wantOK: true,
			want: &Aliaser{
				Aliases: map[string]*Value{
					"salt": stringList("Na", "Cl"),
				},
			},
		},
		// DeleteAlias tests.
		{
			name: "DeleteAlias requires at least one arg",
			a:    &Aliaser{},
			args: []string{"d"},
			wantStderr: []string{
				`no argument provided for "ALIAS"`,
			},
		},
		{
			name: "DeleteAlias handles nonexistent aliases",
			a: &Aliaser{
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
			a: &Aliaser{
				Aliases: map[string]*Value{
					"salt":   stringList("Na", "Cl"),
					"pepper": stringVal("sneezy"),
				},
			},
			args:   []string{"d", "pepper"},
			wantOK: true,
			want: &Aliaser{
				Aliases: map[string]*Value{
					"salt": stringList("Na", "Cl"),
				},
			},
		},
		{
			name: "DeleteAlias handles several args",
			a: &Aliaser{
				Aliases: map[string]*Value{
					"salt":   stringList("Na", "Cl"),
					"pepper": stringVal("sneezy"),
				},
			},
			args:   []string{"d", "garlic", "pepper", "other"},
			wantOK: true,
			want: &Aliaser{
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
			a:    &Aliaser{},
			args: []string{"g"},
			wantStderr: []string{
				`no argument provided for "ALIAS"`,
			},
		},
		{
			name: "GetAlias fails if alias does not exist",
			a:    &Aliaser{},
			args: []string{"g", "pepper"},
			wantStderr: []string{
				`Alias "pepper" does not exist`,
			},
		},
		{
			name: "GetAlias gets an alias",
			a: &Aliaser{
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
			a: &Aliaser{
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
		// SearchAlias searches for aliases.
		{
			name: "SearchAlias requires a regex",
			a:    &Aliaser{},
			args: []string{"s"},
			wantStderr: []string{
				`no argument provided for "REGEXP"`,
			},
		},
		{
			name: "SearchAlias requires a valid regex",
			a:    &Aliaser{},
			args: []string{"s", ":)"},
			wantStderr: []string{
				"Invalid regexp: error parsing regexp: unexpected ): `:)`",
			},
		},
		{
			name: "SearchAlias works",
			a: &Aliaser{
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
	} {
		t.Run(test.name, func(t *testing.T) {
			tcos := &TestCommandOS{}
			got, ok := Execute(tcos, test.a.Command(), test.args, nil)
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
			changed := test.a != nil && test.a.Changed()
			if changed != wantChanged {
				t.Fatalf("Execute(%v) marked Changed as %v; want %v", test.args, changed, wantChanged)
			}

			// Only check diff if we are expecting a change.
			if wantChanged {
				opts := []cmp.Option{
					cmpopts.IgnoreUnexported(Aliaser{}, genericArgs{}, Value{}, StringList{}),
					cmpopts.IgnoreFields(Aliaser{}, "Arg", "Transformer", "Verifier"),
				}
				if diff := cmp.Diff(test.want, test.a, opts...); diff != "" {
					t.Fatalf("Execute(%v) produced emacs diff (-want, +got):\n%s", test.args, diff)
				}
			}
		})
	}
}
