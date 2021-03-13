package commands

// TODO: split this up into separate files (not separate packages).

import (
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestUsage(t *testing.T) {
	for _, test := range []struct {
		name string
		cmd  Command
		want []string
	}{
		{
			name: "returns proper usage",
			cmd:  branchCommand(NoopExecutor, &Completor{}),
			want: []string{
				// TODO: improve this
				"advanced", "first", "\n",
				"foremost", "\n",
				"liszt", "LIST-ARG", "[LIST-ARG ...]", "--inside|-i", "FLAG_VALUE", "FLAG_VALUE", "\n",
				"other", "\n",
				"[", "CB-COMMAND", "CB-COMMAND", "]",
				"\n",
				"basic", "VAL_1", "VARIABLE_2", "--american|-a", "--another", "FLAG_VALUE", "--state|-s", "FLAG_VALUE", "\n",
				"basically", "ANYTHING", "ANYTHING", "ANYTHING", "\n",
				"beginner", "\n", "dquo", "WHOSE", "WHOSE", "\n",
				"ignore", "alpha", "\n", "ayo", "\n", "AIGHT", "\n",
				"intermediate", "SYLLABLE", "SYLLABLE", "SYLLABLE", "--american|-a", "--another", "FLAG_VALUE", "--state|-s", "FLAG_VALUE", "\n",
				"mw", "ALPHA", "ALPHA", "\n",
				"prefixes", "ALPHAS", "\n",
				"sometimes", "OPT_GROUP", "[", "OPT_GROUP", "OPT_GROUP", "OPT_GROUP", "]", "\n",
				"squo", "WHOSE", "WHOSE", "\n",
				"valueTypes",
				"bool", "REQ", "[", "OPT", "]", "--vFlag|-v", "\n",
				"float", "REQ", "[", "OPT", "]", "--vFlag|-v", "FLAG_VALUE", "\n",
				"floatList", "REQ", "REQ", "[", "REQ", "]", "--vFlag|-v", "FLAG_VALUE", "FLAG_VALUE", "[", "FLAG_VALUE", "]", "\n",
				"int", "REQ", "[", "OPT", "]", "--vFlag|-v", "FLAG_VALUE", "\n",
				"intList", "REQ", "REQ", "[", "REQ", "]", "--vFlag|-v", "FLAG_VALUE", "FLAG_VALUE", "[", "FLAG_VALUE", "]", "\n",
				"string", "REQ", "[", "OPT", "]", "--vFlag|-v", "FLAG_VALUE", "\n",
				"stringList", "REQ", "REQ", "[", "REQ", "]", "--vFlag|-v", "FLAG_VALUE", "FLAG_VALUE", "[", "FLAG_VALUE", "]", "\n",
				"\n",
				"wave", "ANY", "ANY", "--yourFlag|-y", "FLAG_VALUE", "FLAG_VALUE", "FLAG_VALUE", "\n",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := test.cmd.Usage()
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("command.Usage() returned diff (-want, +got):\n%s", diff)
			}
		})
	}
}

func TestSet(t *testing.T) {
	for _, test := range []struct {
		name        string
		args        []string
		cArgs       []Arg
		cFlags      []Flag
		wantArgSet  map[string]bool
		wantFlagSet map[string]bool
	}{
		{
			name: "set is true when required string provided",
			args: []string{"hello"},
			cArgs: []Arg{
				StringArg("strArg", true, nil),
			},
			wantArgSet: map[string]bool{
				"strArg": true,
			},
		},
		{
			name: "set is true when optional string provided",
			args: []string{"hello"},
			cArgs: []Arg{
				StringArg("strArg", false, nil),
			},
			wantArgSet: map[string]bool{
				"strArg": true,
			},
		},
		{
			name: "set is false when no optional string provided",
			args: []string{},
			cArgs: []Arg{
				StringArg("strArg", false, nil),
			},
			wantArgSet: map[string]bool{
				"strArg": false,
			},
		},
		{
			name: "set is true when int provided",
			args: []string{"12"},
			cArgs: []Arg{
				IntArg("intArg", false, nil),
			},
			wantArgSet: map[string]bool{
				"intArg": true,
			},
		},
		{
			name: "set is false when int is not provided",
			args: []string{},
			cArgs: []Arg{
				IntArg("intArg", false, nil),
			},
			wantArgSet: map[string]bool{
				"intArg": false,
			},
		},
		{
			name: "set is true when float provided",
			args: []string{"1.2"},
			cArgs: []Arg{
				FloatArg("ftArg", false, nil),
			},
			wantArgSet: map[string]bool{
				"ftArg": true,
			},
		},
		{
			name: "set is false when float is not provided",
			args: []string{},
			cArgs: []Arg{
				FloatArg("ftArg", false, nil),
			},
			wantArgSet: map[string]bool{
				"ftArg": false,
			},
		},
		{
			name: "set is true when bool provided",
			args: []string{"false"},
			cArgs: []Arg{
				BoolArg("bArg", false),
			},
			wantArgSet: map[string]bool{
				"bArg": true,
			},
		},
		{
			name: "set is false when bool is not provided",
			args: []string{},
			cArgs: []Arg{
				BoolArg("bArg", false),
			},
			wantArgSet: map[string]bool{
				"bArg": false,
			},
		},
		{
			name: "set is true when string list provided",
			args: []string{"he", "yo"},
			cArgs: []Arg{
				StringListArg("slArg", 0, 3, nil),
			},
			wantArgSet: map[string]bool{
				"slArg": true,
			},
		},
		{
			name: "set is false when string list is not provided",
			args: []string{},
			cArgs: []Arg{
				StringListArg("slArg", 0, 3, nil),
			},
			wantArgSet: map[string]bool{
				"slArg": false,
			},
		},
		{
			name: "set is true when int list provided",
			args: []string{"2", "-46"},
			cArgs: []Arg{
				IntListArg("ilArg", 0, 3, nil),
			},
			wantArgSet: map[string]bool{
				"ilArg": true,
			},
		},
		{
			name: "set is false when int list is not provided",
			args: []string{},
			cArgs: []Arg{
				IntListArg("ilArg", 0, 3, nil),
			},
			wantArgSet: map[string]bool{
				"ilArg": false,
			},
		},
		{
			name: "set is true when float list provided",
			args: []string{"0.2", "-4.6"},
			cArgs: []Arg{
				FloatListArg("flArg", 0, 3, nil),
			},
			wantArgSet: map[string]bool{
				"flArg": true,
			},
		},
		{
			name: "set is false when float list is not provided",
			args: []string{},
			cArgs: []Arg{
				FloatListArg("flArg", 0, 3, nil),
			},
			wantArgSet: map[string]bool{
				"flArg": false,
			},
		},
		// Flags set
		{
			name: "set is true when string flag provided",
			args: []string{"-f", "hello"},
			cFlags: []Flag{
				StringFlag("strF", 'f', nil),
			},
			wantFlagSet: map[string]bool{
				"strF": true,
			},
		},
		{
			name: "set is false when string flag is not provided",
			args: []string{},
			cFlags: []Flag{
				StringFlag("strF", 'f', nil),
			},
			wantFlagSet: map[string]bool{
				"strF": false,
			},
		},
		{
			name: "set is true when int flag provided",
			args: []string{"-f", "12"},
			cFlags: []Flag{
				IntFlag("intF", 'f', nil),
			},
			wantFlagSet: map[string]bool{
				"intF": true,
			},
		},
		{
			name: "set is false when int flag is not provided",
			args: []string{},
			cFlags: []Flag{
				IntFlag("intF", 'f', nil),
			},
			wantFlagSet: map[string]bool{
				"intF": false,
			},
		},
		{
			name: "set is true when float flag provided",
			args: []string{"-f", "-1.2"},
			cFlags: []Flag{
				FloatFlag("flF", 'f', nil),
			},
			wantFlagSet: map[string]bool{
				"flF": true,
			},
		},
		{
			name: "set is false when float flag is not provided",
			args: []string{},
			cFlags: []Flag{
				FloatFlag("flF", 'f', nil),
			},
			wantFlagSet: map[string]bool{
				"flF": false,
			},
		},
		{
			name: "set is true when bool flag provided",
			args: []string{"-f"},
			cFlags: []Flag{
				BoolFlag("bF", 'f'),
			},
			wantFlagSet: map[string]bool{
				"bF": true,
			},
		},
		{
			name: "set is false when bool flag is not provided",
			args: []string{},
			cFlags: []Flag{
				BoolFlag("bF", 'f'),
			},
			wantFlagSet: map[string]bool{
				"bF": false,
			},
		},
		{
			name: "set is true when string list flag provided",
			args: []string{"-f", "one", "two", "three-four"},
			cFlags: []Flag{
				StringListFlag("slF", 'f', 0, 5, nil),
			},
			wantFlagSet: map[string]bool{
				"slF": true,
			},
		},
		{
			name: "set is false when string list flag not provided",
			args: []string{},
			cFlags: []Flag{
				StringListFlag("slF", 'f', 0, 3, nil),
			},
			wantFlagSet: map[string]bool{
				"slF": false,
			},
		},
		{
			name: "set is true when int list flag provided",
			args: []string{"-f", "12", "34"},
			cFlags: []Flag{
				IntListFlag("ilF", 'f', 0, 3, nil),
			},
			wantFlagSet: map[string]bool{
				"ilF": true,
			},
		},
		{
			name: "set is false when int list flag not provided",
			args: []string{},
			cFlags: []Flag{
				IntListFlag("ilF", 'f', 0, 3, nil),
			},
			wantFlagSet: map[string]bool{
				"ilF": false,
			},
		},
		{
			name: "set is true when float list flag provided",
			args: []string{"-f", "-1.2", "0.34"},
			cFlags: []Flag{
				FloatListFlag("flF", 'f', 0, 3, nil),
			},
			wantFlagSet: map[string]bool{
				"flF": true,
			},
		},
		{
			name: "set is false when float list flag not provided",
			args: []string{},
			cFlags: []Flag{
				FloatListFlag("flF", 'f', 0, 3, nil),
			},
			wantFlagSet: map[string]bool{
				"flF": false,
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			gotArgsSet := map[string]bool{}
			gotFlagsSet := map[string]bool{}
			ex := func(cos CommandOS, args, flags map[string]*Value, _ *OptionInfo) (*ExecutorResponse, bool) {
				for _, a := range test.cArgs {
					gotArgsSet[a.Name()] = args[a.Name()].Provided()
				}
				for _, f := range test.cFlags {
					gotFlagsSet[f.Name()] = flags[f.Name()].Provided()
				}
				return nil, true
			}

			c := &TerminusCommand{
				Executor: ex,
				Args:     test.cArgs,
				Flags:    test.cFlags,
			}
			tcos := &TestCommandOS{}
			if _, ok := Execute(tcos, c, test.args, nil); !ok {
				t.Fatalf("commands.Execute(%s) failed: %v", test.args, tcos)
			}

			if len(gotArgsSet) == 0 {
				gotArgsSet = nil
			}
			if len(gotFlagsSet) == 0 {
				gotFlagsSet = nil
			}
			if diff := cmp.Diff(test.wantArgSet, gotArgsSet); diff != "" {
				t.Errorf("commands.Execute(%v) had improperly set args: \n%s", test.args, diff)
			}
			if diff := cmp.Diff(test.wantFlagSet, gotFlagsSet); diff != "" {
				t.Errorf("commands.Execute(%v) had improperly set flags: \n%s", test.args, diff)
			}
		})
	}
}

func branchCommand(executor Executor, completor *Completor, opts ...ArgOpt) Command {
	return &CommandBranch{
		Subcommands: map[string]Command{
			"advanced": &CommandBranch{
				TerminusCommand: &TerminusCommand{
					Executor: executor,
					Args: []Arg{
						StringListArg("cb-command", 0, 2, completor, opts...),
					},
				},
				Subcommands: map[string]Command{
					"first": &TerminusCommand{
						Executor: executor,
					},
					"foremost": &CommandBranch{},
					"other":    &TerminusCommand{},
					"liszt": &TerminusCommand{
						Executor: executor,
						Args: []Arg{
							StringListArg("list-arg", 1, UnboundedList, completor, opts...),
						},
						Flags: []Flag{
							StringListFlag("inside", 'i', 2, 0, completor, opts...),
						},
					},
				},
			},
			"basic": &TerminusCommand{
				Executor: executor,
				Args: []Arg{
					StringListArg("val_1", 1, 0, completor, opts...),
					StringListArg("variable 2", 1, 0, completor, opts...),
				},
				Flags: []Flag{
					BoolFlag("american", 'a', opts...),
					StringListFlag("another", 0, 1, 0, completor, opts...),
					StringListFlag("state", 's', 1, 0, completor, opts...),
				},
			},
			"basically": &TerminusCommand{
				Args: []Arg{
					// completor is explicitly nil for a test.
					StringListArg("anything", 3, 0, nil, opts...),
				},
			},
			"beginner": &CommandBranch{},
			"intermediate": &TerminusCommand{
				Executor: executor,
				Args: []Arg{
					StringListArg("syllable", 3, 0, completor, opts...),
				},
				Flags: []Flag{
					BoolFlag("american", 'a', opts...),
					StringListFlag("another", 0, 1, 0, completor, opts...),
					StringListFlag("state", 's', 1, 0, completor, opts...),
				},
			},
			"sometimes": &TerminusCommand{
				Executor: executor,
				Args: []Arg{
					StringListArg("opt group", 1, 3, completor, opts...),
				},
			},
			"prefixes": &TerminusCommand{
				Args: []Arg{
					StringListArg("alphas", 1, 0, completor, opts...),
				},
			},
			"squo": &TerminusCommand{
				Args: []Arg{
					StringListArg("whose", 2, 0, completor, opts...),
				},
			},
			"dquo": &TerminusCommand{
				Args: []Arg{
					StringListArg("whose", 2, 0, completor, opts...),
				},
			},
			"mw": &TerminusCommand{
				Args: []Arg{
					StringListArg("alpha", 2, 0, completor, opts...),
				},
			},
			"wave": &TerminusCommand{
				Args: []Arg{
					StringListArg("any", 2, 0, completor, opts...),
				},
				Flags: []Flag{
					StringListFlag("yourFlag", 'y', 3, 0, completor, opts...),
				},
			},
			"ignore": &CommandBranch{
				IgnoreSubcommandAutocomplete: true,
				Subcommands: map[string]Command{
					"alpha": &TerminusCommand{},
					"ayo":   &TerminusCommand{},
				},
				TerminusCommand: &TerminusCommand{
					Args: []Arg{
						StringArg("aight", true, completor, opts...),
					},
				},
			},
			"valueTypes": &CommandBranch{
				Subcommands: map[string]Command{
					"string": &TerminusCommand{
						Executor: executor,
						Args: []Arg{
							StringArg("req", true, completor, opts...),
							StringArg("opt", false, completor, opts...),
						},
						Flags: []Flag{
							StringFlag("vFlag", 'v', completor, opts...),
						},
					},
					"stringList": &TerminusCommand{
						Executor: executor,
						Args: []Arg{
							StringListArg("req", 2, 1, completor, opts...),
						},
						Flags: []Flag{
							StringListFlag("vFlag", 'v', 2, 1, completor, opts...),
						},
					},
					"int": &TerminusCommand{
						Executor: executor,
						Args: []Arg{
							IntArg("req", true, completor, opts...),
							IntArg("opt", false, completor, opts...),
						},
						Flags: []Flag{
							IntFlag("vFlag", 'v', completor, opts...),
						},
					},
					"intList": &TerminusCommand{
						Executor: executor,
						Args: []Arg{
							IntListArg("req", 2, 1, completor, opts...),
						},
						Flags: []Flag{
							IntListFlag("vFlag", 'v', 2, 1, completor, opts...),
						},
					},
					"float": &TerminusCommand{
						Executor: executor,
						Args: []Arg{
							FloatArg("req", true, completor, opts...),
							FloatArg("opt", false, completor, opts...),
						},
						Flags: []Flag{
							FloatFlag("vFlag", 'v', completor, opts...),
						},
					},
					"floatList": &TerminusCommand{
						Executor: executor,
						Args: []Arg{
							FloatListArg("req", 2, 1, completor, opts...),
						},
						Flags: []Flag{
							FloatListFlag("vFlag", 'v', 2, 1, completor, opts...),
						},
					},
					"bool": &TerminusCommand{
						Executor: executor,
						Args: []Arg{
							BoolArg("req", true, opts...),
							BoolArg("opt", false, opts...),
						},
						Flags: []Flag{
							BoolFlag("vFlag", 'v', opts...),
						},
					},
				},
			},
		},
	}
}

func TestExecute(t *testing.T) {
	for _, test := range []struct {
		name             string
		args             []string
		ex               Executor
		exResp           *ExecutorResponse
		opts             []ArgOpt
		want             *ExecutorResponse
		wantStderr       []string
		wantStdout       []string
		wantExecuteArgs  map[string]*Value
		wantExecuteFlags map[string]*Value
		wantOK           bool
	}{
		// Basic tests
		{
			name:       "empty args",
			wantStderr: []string{"more args required"},
		},
		{
			name:       "incomplete command",
			args:       []string{"huh"},
			wantStderr: []string{`unknown subcommand and no terminus command defined`},
		},
		{
			name:       "not enough flag values",
			args:       []string{"basic", "--state"},
			wantStderr: []string{`failed to process flags: not enough arguments`},
		},
		{
			name:       "too many positional arguments",
			args:       []string{"basic", "--state", "maine", "build", "one", "else", "too"},
			wantStderr: []string{"extra unknown args ([else too])"},
		},
		{
			name:       "not enough positional arguments",
			args:       []string{"intermediate", "--state", "maine", "one"},
			wantStderr: []string{`failed to process args: not enough arguments`},
		},
		{
			name:       "not enough positional arguments",
			args:       []string{"basic", "--state", "maine"},
			wantStderr: []string{`no argument provided for "val_1"`},
		},
		{
			name:       "no executor defined",
			args:       []string{"advanced", "other"},
			wantStderr: []string{"no executor defined for command"},
		},
		{
			name:   "works when CommandBranch defines terminusCommand",
			args:   []string{"advanced", "not", "registered"},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"cb-command": StringListValue("not", "registered"),
			},
		},
		{
			name: "fails when CommandBranch defines executor fails",
			ex: func(cos CommandOS, args, flags map[string]*Value, _ *OptionInfo) (*ExecutorResponse, bool) {
				cos.Stderr("bad news bears")
				return nil, false
			},
			args:       []string{"advanced", "not", "registered"},
			wantStderr: []string{"bad news bears"},
		},
		{
			name:   "works with no flags",
			args:   []string{"basic", "un", "deux"},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"val_1":      StringListValue("un"),
				"variable 2": StringListValue("deux"),
			},
		},
		{
			name:   "works with flags at the beginning",
			args:   []string{"basic", "--state", "jersey", "trois", "quatre"},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"val_1":      StringListValue("trois"),
				"variable 2": StringListValue("quatre"),
			},
			wantExecuteFlags: map[string]*Value{
				"state": StringListValue("jersey"),
			},
		},
		{
			name:   "works with flags in the middle",
			args:   []string{"basic", "trois", "--state", "massachusetts", "quatre"},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"val_1":      StringListValue("trois"),
				"variable 2": StringListValue("quatre"),
			},
			wantExecuteFlags: map[string]*Value{
				"state": StringListValue("massachusetts"),
			},
		},
		{
			name:   "works with flags at the end",
			args:   []string{"basic", "trois", "quatre", "-s", "connecticut"},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"val_1":      StringListValue("trois"),
				"variable 2": StringListValue("quatre"),
			},
			wantExecuteFlags: map[string]*Value{
				"state": StringListValue("connecticut"),
			},
		},
		{
			name:   "works with boolean flag",
			args:   []string{"basic", "trois", "--american", "quatre"},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"val_1":      StringListValue("trois"),
				"variable 2": StringListValue("quatre"),
			},
			wantExecuteFlags: map[string]*Value{
				"american": BoolValue(true),
			},
		},
		{
			name:   "works with short boolean flag",
			args:   []string{"basic", "-a", "trois", "quatre"},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"val_1":      StringListValue("trois"),
				"variable 2": StringListValue("quatre"),
			},
			wantExecuteFlags: map[string]*Value{
				"american": BoolValue(true),
			},
		},
		{
			name:   "works with arguments with multiple args",
			args:   []string{"intermediate", "first", "2nd", "bronze"},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"syllable": StringListValue("first", "2nd", "bronze"),
			},
		},
		// Test lists
		{
			name:       "list fails when not enough args",
			args:       []string{"advanced", "liszt"},
			wantStderr: []string{`no argument provided for "list-arg"`},
		},
		{
			name:   "list succeeds when at minimum args",
			args:   []string{"advanced", "liszt", "piano"},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"list-arg": StringListValue("piano"),
			},
		},
		{
			name:   "list succeeds when extra args",
			args:   []string{"advanced", "liszt", "piano", "harp", "picolo"},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"list-arg": StringListValue("piano", "harp", "picolo"),
			},
		},
		{
			name:   "list succeeds when flag in between",
			args:   []string{"advanced", "liszt", "piano", "--inside", "56", "34", "harp", "picolo"},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"list-arg": StringListValue("piano", "harp", "picolo"),
			},
			wantExecuteFlags: map[string]*Value{
				"inside": StringListValue("56", "34"),
			},
		},
		{
			name:   "list succeeds when short flag in between",
			args:   []string{"advanced", "liszt", "piano", "-i", "56", "34", "harp", "picolo"},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"list-arg": StringListValue("piano", "harp", "picolo"),
			},
			wantExecuteFlags: map[string]*Value{
				"inside": StringListValue("56", "34"),
			},
		},
		// Test extra optional arguments.
		{
			name:       "optional argument doesn't accept less than minimum",
			args:       []string{"sometimes"},
			wantStderr: []string{`no argument provided for "opt group"`},
		},
		{
			name:   "optional argument accepts minimum",
			args:   []string{"sometimes", "temp"},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"opt group": StringListValue("temp"),
			},
		},
		{
			name:   "optional argument accepts middle amount",
			args:   []string{"sometimes", "temp", "occasional"},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"opt group": StringListValue("temp", "occasional"),
			},
		},
		{
			name:   "optional argument accepts max amount",
			args:   []string{"sometimes", "temp", "occasional", "tmp", "temporary"},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"opt group": StringListValue("temp", "occasional", "tmp", "temporary"),
			},
		},
		{
			name:       "optional argument does not accept more than max amount",
			args:       []string{"sometimes", "temp", "occasional", "tmp", "temporary", "occ"},
			wantStderr: []string{"extra unknown args ([occ])"},
		},
		// Test return values
		{
			name:   "returns what the executor returns",
			args:   []string{"intermediate", "first", "2nd", "bronze"},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"syllable": StringListValue("first", "2nd", "bronze"),
			},
			exResp: &ExecutorResponse{Executable: []string{"this", "was a", "success"}},
			want:   &ExecutorResponse{Executable: []string{"this", "was a", "success"}},
		},
		{
			name: "fails when executor returns false",
			args: []string{"intermediate", "first", "2nd", "bronze"},
			ex: func(cos CommandOS, args, flags map[string]*Value, _ *OptionInfo) (*ExecutorResponse, bool) {
				cos.Stderr("this was a failure")
				return nil, false
			},
			wantStderr: []string{"this was a failure"},
		},
		// CommandBranch with terminus command
		{
			name:   "branch command's terminus command with no arguments",
			args:   []string{"advanced"},
			wantOK: true,
		},
		// Commands with different value types.
		// string argument type
		{
			name:   "handles string argument",
			args:   []string{"valueTypes", "string", "hello"},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": StringValue("hello"),
			},
		},
		{
			name:   "handles optional string argument",
			args:   []string{"valueTypes", "string", "hello", "there"},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": StringValue("hello"),
				"opt": StringValue("there"),
			},
		},
		// stringList argument type
		{
			name:   "handles stringList argument",
			args:   []string{"valueTypes", "stringList", "its", "me"},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": StringListValue("its", "me"),
			},
		},
		{
			name:   "handles optional stringList arguments",
			args:   []string{"valueTypes", "stringList", "its", "me", "mario"},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": StringListValue("its", "me", "mario"),
			},
		},
		// int argument type
		{
			name:   "handles int argument",
			args:   []string{"valueTypes", "int", "123"},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": IntValue(123),
			},
		},
		{
			name:   "handles optional int argument",
			args:   []string{"valueTypes", "int", "123", "-45"},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": IntValue(123),
				"opt": IntValue(-45),
			},
		},
		{
			name:       "int argument requires int value",
			args:       []string{"valueTypes", "int", "123.45"},
			wantStderr: []string{`failed to process args: argument should be an integer: strconv.Atoi: parsing "123.45": invalid syntax`},
		},
		{
			name:       "int flag requires int value",
			args:       []string{"valueTypes", "int", "-v", "123.45"},
			wantStderr: []string{`failed to process flags: argument should be an integer: strconv.Atoi: parsing "123.45": invalid syntax`},
		},
		// intList argument type
		{
			name:   "handles intList argument",
			args:   []string{"valueTypes", "intList", "123", "-45"},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": IntListValue(123, -45),
			},
		},
		{
			name:   "handles optional intList arguments",
			args:   []string{"valueTypes", "intList", "123", "-45", "0"},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": IntListValue(123, -45, 0),
			},
		},
		{
			name:       "int list argument requires int values",
			args:       []string{"valueTypes", "intList", "-10", "123.45"},
			wantStderr: []string{`failed to process args: strconv.Atoi: parsing "123.45": invalid syntax`},
		},
		{
			name:       "int list argument requires int values#2",
			args:       []string{"valueTypes", "intList", "-v", "123.45", "67"},
			wantStderr: []string{`failed to process flags: strconv.Atoi: parsing "123.45": invalid syntax`},
		},
		// float argument type
		{
			name:   "handles float argument",
			args:   []string{"valueTypes", "float", "123.45"},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": FloatValue(123.45),
			},
		},
		{
			name:   "handles optional float argument",
			args:   []string{"valueTypes", "float", "123.45", "-67"},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": FloatValue(123.45),
				"opt": FloatValue(-67),
			},
		},
		{
			name:       "float argument requires float value",
			args:       []string{"valueTypes", "float", "twelve"},
			wantStderr: []string{`failed to process args: argument should be a float: strconv.ParseFloat: parsing "twelve": invalid syntax`},
		},
		{
			name:       "float flag requires float value",
			args:       []string{"valueTypes", "float", "--vFlag", "twelve"},
			wantStderr: []string{`failed to process flags: argument should be a float: strconv.ParseFloat: parsing "twelve": invalid syntax`},
		},
		// floatList argument type
		{
			name:   "handles floatList argument",
			args:   []string{"valueTypes", "floatList", "123.45", "-67"},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": FloatListValue(123.45, -67),
			},
		},
		{
			name:   "handles optional floatList arguments",
			args:   []string{"valueTypes", "floatList", "123.45", "-67", "0"},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": FloatListValue(123.45, -67, 0),
			},
		},
		{
			name:       "float list argument requires float values",
			args:       []string{"valueTypes", "floatList", "-10", "twelve"},
			wantStderr: []string{`failed to process args: strconv.ParseFloat: parsing "twelve": invalid syntax`},
		},
		{
			name:       "float list flag requires float values",
			args:       []string{"valueTypes", "floatList", "-v", "3.5", "twelve"},
			wantStderr: []string{`failed to process flags: strconv.ParseFloat: parsing "twelve": invalid syntax`},
		},
		// bool argument type
		{
			name:   "handles bool argument",
			args:   []string{"valueTypes", "bool", "true"},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": BoolValue(true),
			},
		},
		{
			name:   "handles optional bool argument",
			args:   []string{"valueTypes", "bool", "false", "true"},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": BoolValue(false),
				"opt": BoolValue(true),
			},
		},
		{
			name:   "allows shorthand bool argument",
			args:   []string{"valueTypes", "bool", "t", "f"},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": BoolValue(true),
				"opt": BoolValue(false),
			},
		},
		{
			name:       "bool argument requires bool value",
			args:       []string{"valueTypes", "bool", "maybe"},
			wantStderr: []string{`failed to process args: argument should be a bool: strconv.ParseBool: parsing "maybe": invalid syntax`},
		},
		{
			name:   "bool flag works",
			args:   []string{"valueTypes", "bool", "--vFlag", "false"},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": BoolValue(false),
			},
			wantExecuteFlags: map[string]*Value{
				"vFlag": BoolValue(true),
			},
		},
		{
			name:   "bool shorthand flag works",
			args:   []string{"valueTypes", "bool", "-v", "true"},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": BoolValue(true),
			},
			wantExecuteFlags: map[string]*Value{
				"vFlag": BoolValue(true),
			},
		},
		// ArgOpt tests
		{
			name: "Breaks when arg option is for invalid type",
			args: []string{"valueTypes", "string", "123"},
			opts: []ArgOpt{
				IntEQ(123),
			},
			wantStderr: []string{"failed to process args: option can only be bound to arguments with type 3"},
		},
		// Contains
		{
			name: "Contains works",
			args: []string{"valueTypes", "string", "goodbye"},
			opts: []ArgOpt{
				Contains("good"),
			},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": StringValue("goodbye"),
			},
		},
		{
			name: "Contains fails",
			args: []string{"valueTypes", "string", "hello"},
			opts: []ArgOpt{
				Contains("good"),
			},
			wantStderr: []string{`failed to process args: validation failed: [Contains] value doesn't contain substring "good"`},
		},
		// MinLength
		{
			name: "MinLength fails if too few characters",
			args: []string{"valueTypes", "string", "ab"},
			opts: []ArgOpt{
				MinLength(3),
			},
			wantStderr: []string{`failed to process args: validation failed: [MinLength] value must be at least 3 characters`},
		},
		{
			name: "MinLength passes when exact number of characters",
			args: []string{"valueTypes", "string", "abc"},
			opts: []ArgOpt{
				MinLength(3),
			},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": StringValue("abc"),
			},
		},
		{
			name: "MinLength passes when exact number of characters",
			args: []string{"valueTypes", "string", "abcd"},
			opts: []ArgOpt{
				MinLength(3),
			},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": StringValue("abcd"),
			},
		},
		// IntEQ
		{
			name: "IntEQ works",
			args: []string{"valueTypes", "int", "24"},
			opts: []ArgOpt{
				IntEQ(24),
			},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": IntValue(24),
			},
		},
		{
			name: "IntEQ fails",
			args: []string{"valueTypes", "int", "25"},
			opts: []ArgOpt{
				IntEQ(24),
			},
			wantStderr: []string{"failed to process args: validation failed: [IntEQ] value isn't equal to 24"},
		},
		// IntNE
		{
			name: "IntNE works",
			args: []string{"valueTypes", "int", "24"},
			opts: []ArgOpt{
				IntNE(25),
			},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": IntValue(24),
			},
		},
		{
			name: "IntNE fails",
			args: []string{"valueTypes", "int", "25"},
			opts: []ArgOpt{
				IntNE(25),
			},
			wantStderr: []string{"failed to process args: validation failed: [IntNE] value isn't not equal to 25"},
		},
		// IntLT
		{
			name: "IntLT works",
			args: []string{"valueTypes", "int", "24"},
			opts: []ArgOpt{
				IntLT(25),
			},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": IntValue(24),
			},
		},
		{
			name: "IntLT fails when equal",
			args: []string{"valueTypes", "int", "25"},
			opts: []ArgOpt{
				IntLT(25),
			},
			wantStderr: []string{"failed to process args: validation failed: [IntLT] value isn't less than 25"},
		},
		{
			name: "IntLT fails when not less",
			args: []string{"valueTypes", "int", "26"},
			opts: []ArgOpt{
				IntLT(25),
			},
			wantStderr: []string{"failed to process args: validation failed: [IntLT] value isn't less than 25"},
		},
		// IntLTE
		{
			name: "IntLTE works",
			args: []string{"valueTypes", "int", "24"},
			opts: []ArgOpt{
				IntLTE(25),
			},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": IntValue(24),
			},
		},
		{
			name: "IntLTE works when equal",
			args: []string{"valueTypes", "int", "25"},
			opts: []ArgOpt{
				IntLTE(25),
			},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": IntValue(25),
			},
		},
		{
			name: "IntLT fails when not less",
			args: []string{"valueTypes", "int", "26"},
			opts: []ArgOpt{
				IntLTE(25),
			},
			wantStderr: []string{"failed to process args: validation failed: [IntLTE] value isn't less than or equal to 25"},
		},
		// IntLT
		{
			name: "IntGT fails when not greater",
			args: []string{"valueTypes", "int", "24"},
			opts: []ArgOpt{
				IntGT(25),
			},
			wantStderr: []string{"failed to process args: validation failed: [IntGT] value isn't greater than 25"},
		},
		{
			name: "IntGT fails when equal",
			args: []string{"valueTypes", "int", "25"},
			opts: []ArgOpt{
				IntGT(25),
			},
			wantStderr: []string{"failed to process args: validation failed: [IntGT] value isn't greater than 25"},
		},
		{
			name: "IntGT works",
			args: []string{"valueTypes", "int", "26"},
			opts: []ArgOpt{
				IntGT(25),
			},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": IntValue(26),
			},
		},
		// IntGTE
		{
			name: "IntGTE fails when not greater",
			args: []string{"valueTypes", "int", "24"},
			opts: []ArgOpt{
				IntGTE(25),
			},
			wantStderr: []string{"failed to process args: validation failed: [IntGTE] value isn't greater than or equal to 25"},
		},
		{
			name: "IntGTE works when equal",
			args: []string{"valueTypes", "int", "25"},
			opts: []ArgOpt{
				IntGTE(25),
			},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": IntValue(25),
			},
		},
		{
			name: "IntGTE works",
			args: []string{"valueTypes", "int", "26"},
			opts: []ArgOpt{
				IntGTE(25),
			},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": IntValue(26),
			},
		},
		// IntPositive
		{
			name: "IntPositive fails when negative",
			args: []string{"valueTypes", "int", "-1"},
			opts: []ArgOpt{
				IntPositive(),
			},
			wantStderr: []string{"failed to process args: validation failed: [IntPositive] value isn't positive"},
		},
		{
			name: "IntPositive fails when zero",
			args: []string{"valueTypes", "int", "0"},
			opts: []ArgOpt{
				IntPositive(),
			},
			wantStderr: []string{"failed to process args: validation failed: [IntPositive] value isn't positive"},
		},
		{
			name: "IntPositive works when positive",
			args: []string{"valueTypes", "int", "1"},
			opts: []ArgOpt{
				IntPositive(),
			},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": IntValue(1),
			},
		},
		// IntNegative
		{
			name: "IntNegative works when negative",
			args: []string{"valueTypes", "int", "-1"},
			opts: []ArgOpt{
				IntNegative(),
			},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": IntValue(-1),
			},
		},
		{
			name: "IntNegative fails when zero",
			args: []string{"valueTypes", "int", "0"},
			opts: []ArgOpt{
				IntNegative(),
			},
			wantStderr: []string{"failed to process args: validation failed: [IntNegative] value isn't negative"},
		},
		{
			name: "IntNegative fails when positive",
			args: []string{"valueTypes", "int", "1"},
			opts: []ArgOpt{
				IntNegative(),
			},
			wantStderr: []string{"failed to process args: validation failed: [IntNegative] value isn't negative"},
		},
		// IntNonNegative
		{
			name: "IntNonNegative fails when negative",
			args: []string{"valueTypes", "int", "-1"},
			opts: []ArgOpt{
				IntNonNegative(),
			},
			wantStderr: []string{"failed to process args: validation failed: [IntNonNegative] value isn't non-negative"},
		},
		{
			name: "IntNonNegative works when zero",
			args: []string{"valueTypes", "int", "0"},
			opts: []ArgOpt{
				IntNonNegative(),
			},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": IntValue(0),
			},
		},
		{
			name: "IntNonNegative works when positive",
			args: []string{"valueTypes", "int", "1"},
			opts: []ArgOpt{
				IntNonNegative(),
			},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": IntValue(1),
			},
		},
		// FloatEQ
		{
			name: "FloatEQ works",
			args: []string{"valueTypes", "float", "24"},
			opts: []ArgOpt{
				FloatEQ(24),
			},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": FloatValue(24),
			},
		},
		{
			name: "FloatEQ fails",
			args: []string{"valueTypes", "float", "25"},
			opts: []ArgOpt{
				FloatEQ(24),
			},
			wantStderr: []string{"failed to process args: validation failed: [FloatEQ] value isn't equal to 24.00"},
		},
		// FloatNE
		{
			name: "FloatNE works",
			args: []string{"valueTypes", "float", "24"},
			opts: []ArgOpt{
				FloatNE(25),
			},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": FloatValue(24),
			},
		},
		{
			name: "FloatNE fails",
			args: []string{"valueTypes", "float", "25"},
			opts: []ArgOpt{
				FloatNE(25),
			},
			wantStderr: []string{"failed to process args: validation failed: [FloatNE] value isn't not equal to 25.00"},
		},
		// FloatLT
		{
			name: "FloatLT works",
			args: []string{"valueTypes", "float", "24"},
			opts: []ArgOpt{
				FloatLT(25),
			},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": FloatValue(24),
			},
		},
		{
			name: "FloatLT fails when equal",
			args: []string{"valueTypes", "float", "25"},
			opts: []ArgOpt{
				FloatLT(25),
			},
			wantStderr: []string{"failed to process args: validation failed: [FloatLT] value isn't less than 25.00"},
		},
		{
			name: "FloatLT fails when not less",
			args: []string{"valueTypes", "float", "26"},
			opts: []ArgOpt{
				FloatLT(25),
			},
			wantStderr: []string{"failed to process args: validation failed: [FloatLT] value isn't less than 25.00"},
		},
		// FloatLTE
		{
			name: "FloatLTE works",
			args: []string{"valueTypes", "float", "24"},
			opts: []ArgOpt{
				FloatLTE(25),
			},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": FloatValue(24),
			},
		},
		{
			name: "FloatLTE works when equal",
			args: []string{"valueTypes", "float", "25"},
			opts: []ArgOpt{
				FloatLTE(25),
			},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": FloatValue(25),
			},
		},
		{
			name: "FloatLT fails when not less",
			args: []string{"valueTypes", "float", "26"},
			opts: []ArgOpt{
				FloatLTE(25),
			},
			wantStderr: []string{"failed to process args: validation failed: [FloatLTE] value isn't less than or equal to 25.00"},
		},
		// FloatGT
		{
			name: "FloatGT fails when not greater",
			args: []string{"valueTypes", "float", "24"},
			opts: []ArgOpt{
				FloatGT(25),
			},
			wantStderr: []string{"failed to process args: validation failed: [FloatGT] value isn't greater than 25.00"},
		},
		{
			name: "FloatGT fails when equal",
			args: []string{"valueTypes", "float", "25"},
			opts: []ArgOpt{
				FloatGT(25),
			},
			wantStderr: []string{"failed to process args: validation failed: [FloatGT] value isn't greater than 25.00"},
		},
		{
			name: "FloatGT works",
			args: []string{"valueTypes", "float", "26"},
			opts: []ArgOpt{
				FloatGT(25),
			},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": FloatValue(26),
			},
		},
		// FloatGTE
		{
			name: "FloatGTE fails when not greater",
			args: []string{"valueTypes", "float", "24"},
			opts: []ArgOpt{
				FloatGTE(25),
			},
			wantStderr: []string{"failed to process args: validation failed: [FloatGTE] value isn't greater than or equal to 25.00"},
		},
		{
			name: "FloatGTE works when equal",
			args: []string{"valueTypes", "float", "25"},
			opts: []ArgOpt{
				FloatGTE(25),
			},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": FloatValue(25),
			},
		},
		{
			name: "FloatGTE works",
			args: []string{"valueTypes", "float", "26"},
			opts: []ArgOpt{
				FloatGTE(25),
			},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": FloatValue(26),
			},
		},
		// FloatPositive
		{
			name: "FloatPositive fails when negative",
			args: []string{"valueTypes", "float", "-1"},
			opts: []ArgOpt{
				FloatPositive(),
			},
			wantStderr: []string{"failed to process args: validation failed: [FloatPositive] value isn't positive"},
		},
		{
			name: "FloatPositive fails when zero",
			args: []string{"valueTypes", "float", "0"},
			opts: []ArgOpt{
				FloatPositive(),
			},
			wantStderr: []string{"failed to process args: validation failed: [FloatPositive] value isn't positive"},
		},
		{
			name: "FloatPositive works when positive",
			args: []string{"valueTypes", "float", "1"},
			opts: []ArgOpt{
				FloatPositive(),
			},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": FloatValue(1),
			},
		},
		// FloatNegative
		{
			name: "FloatNegative works when negative",
			args: []string{"valueTypes", "float", "-1"},
			opts: []ArgOpt{
				FloatNegative(),
			},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": FloatValue(-1),
			},
		},
		{
			name: "FloatNegative fails when zero",
			args: []string{"valueTypes", "float", "0"},
			opts: []ArgOpt{
				FloatNegative(),
			},
			wantStderr: []string{"failed to process args: validation failed: [FloatNegative] value isn't negative"},
		},
		{
			name: "FloatNegative fails when positive",
			args: []string{"valueTypes", "float", "1"},
			opts: []ArgOpt{
				FloatNegative(),
			},
			wantStderr: []string{"failed to process args: validation failed: [FloatNegative] value isn't negative"},
		},
		// FloatNonNegative
		{
			name: "FloatNonNegative fails when negative",
			args: []string{"valueTypes", "float", "-1"},
			opts: []ArgOpt{
				FloatNonNegative(),
			},
			wantStderr: []string{"failed to process args: validation failed: [FloatNonNegative] value isn't non-negative"},
		},
		{
			name: "FloatNonNegative works when zero",
			args: []string{"valueTypes", "float", "0"},
			opts: []ArgOpt{
				FloatNonNegative(),
			},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": FloatValue(0),
			},
		},
		{
			name: "FloatNonNegative works when positive",
			args: []string{"valueTypes", "float", "1"},
			opts: []ArgOpt{
				FloatNonNegative(),
			},
			wantOK: true,
			wantExecuteArgs: map[string]*Value{
				"req": FloatValue(1),
			},
		},
		/* Useful comment for commenting out tests */
	} {
		t.Run(test.name, func(t *testing.T) {
			var gotExecuteArgs map[string]*Value
			var gotExecuteFlags map[string]*Value

			ex := test.ex
			if ex == nil {
				// TODO: verify oi is correct
				ex = func(cos CommandOS, args, flags map[string]*Value, _ *OptionInfo) (*ExecutorResponse, bool) {
					// Check length so we can consider empty to be the same as nil.
					// That makes for cleaner test cases.
					if len(args) > 0 {
						gotExecuteArgs = args
					}
					if len(flags) > 0 {
						gotExecuteFlags = flags
					}
					return test.exResp, true
				}
			}

			cmd := branchCommand(ex, &Completor{}, test.opts...)

			tcos := &TestCommandOS{}

			got, ok := Execute(tcos, cmd, test.args, nil)
			if ok != test.wantOK {
				t.Errorf("commands.Execute(%v) returned %v for ok; want %v", test.args, ok, test.wantOK)
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("command.Execute(%v) returned diff (-want, +got):\n%s", test.args, diff)
			}

			if diff := cmp.Diff(test.wantStdout, tcos.GetStdout()); diff != "" {
				t.Errorf("command.Execute(%v) produced stdout diff (-want, +got):\n%s", test.args, diff)
			}
			if diff := cmp.Diff(test.wantStderr, tcos.GetStderr()); diff != "" {
				t.Errorf("command.Execute(%v) produced stderr diff (-want, +got):\n%s", test.args, diff)
			}

			if diff := cmp.Diff(test.wantExecuteArgs, gotExecuteArgs); diff != "" {
				t.Errorf("command.Execute(%v) produced execute args diff (-want, +got):\n%s", test.args, diff)
			}

			if diff := cmp.Diff(test.wantExecuteFlags, gotExecuteFlags); diff != "" {
				t.Errorf("command.Execute(%v) produced execute flags diff (-want, +got):\n%s", test.args, diff)
			}
		})
	}
}

func TestCommandComplete(t *testing.T) {
	// Note: this uses cmd.Command (as opposed to the package level function "Autocomplete").
	// All more generic tests should go in "TestAutocomplete".
	for _, test := range []struct {
		name string
		cmd  Command
		args []string
		want []string
	}{
		{
			name: "handles nil args",
			cmd: &CommandBranch{
				TerminusCommand: &TerminusCommand{},
				Subcommands: map[string]Command{
					"a": &TerminusCommand{},
					"b": &CommandBranch{},
					"c": &TerminusCommand{},
				},
			},
			want: []string{"a", "b", "c"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			suggestions := test.cmd.Complete(test.args).Suggestions
			sort.Strings(suggestions)
			if diff := cmp.Diff(test.want, suggestions); diff != "" {
				t.Errorf("Complete(%v) produced diff (-want, +got):\n%s", test.args, diff)
			}
		})
	}
}

func TestAutocomplete(t *testing.T) {
	for _, test := range []struct {
		name              string
		cmd               Command
		args              []string
		cursorIdx         int
		distinct          bool
		want              []string
		fetchResp         []string
		wantCompleteArgs  map[string]*Value
		wantCompleteFlags map[string]*Value
		wantValue         *Value
	}{
		// Basic tests
		{
			name: "nil arg gets predicted",
			want: []string{
				"advanced",
				"basic",
				"basically",
				"beginner",
				"dquo",
				"ignore",
				"intermediate",
				"mw",
				"prefixes",
				"sometimes",
				"squo",
				"valueTypes",
				"wave",
			},
		},
		{
			name: "empty arg gets predicted",
			args: []string{""},
			want: []string{
				"advanced",
				"basic",
				"basically",
				"beginner",
				"dquo",
				"ignore",
				"intermediate",
				"mw",
				"prefixes",
				"sometimes",
				"squo",
				"valueTypes",
				"wave",
			},
		},
		{
			name: "partially complete arg gets multiple recommendations",
			args: []string{"b"},
			want: []string{"basic", "basically", "beginner"},
		},
		{
			name: "partially complete arg gets multiple recommendations when a match is subset",
			args: []string{"basic"},
			want: []string{"basic", "basically"},
		},
		{
			name: "completes when exact macth with one option",
			args: []string{"advanced"},
			want: []string{"advanced"},
		},
		{
			name: "no complete when unknown arg",
			args: []string{"unknown"},
		},
		{
			name:      "empty second arg gets autocompleted",
			args:      []string{"basic", ""},
			fetchResp: []string{"build", "test", "try", "trying"},
			want:      []string{"build", "test", "try", "trying"},
			wantValue: StringListValue(""),
			wantCompleteArgs: map[string]*Value{
				"val_1": StringListValue(""),
			},
		},
		{
			name:      "partially complete second arg gets autocompleted",
			args:      []string{"basic", "t"},
			fetchResp: []string{"build", "test", "try", "trying"},
			want:      []string{"test", "try", "trying"},
			wantValue: StringListValue("t"),
			wantCompleteArgs: map[string]*Value{
				"val_1": StringListValue("t"),
			},
		},
		{
			name: "no autocomplete when run out of defined args",
			args: []string{"basic", "one", "two", ""},
		},
		{
			name:      "works with more than two argument groups",
			args:      []string{"basic", "build", "o"},
			fetchResp: []string{"one", "other", "value"},
			want:      []string{"one", "other"},
			wantValue: StringListValue("o"),
			wantCompleteArgs: map[string]*Value{
				"val_1":      StringListValue("build"),
				"variable 2": StringListValue("o"),
			},
		},
		{
			name: "nil completor has no autocomplete",
			args: []string{"basically", ""},
		},
		// Ignores subcommand autocompletes
		{
			name:      "IgnoreSubcommandAutocomplete doesn't include subcommands in suggestions",
			args:      []string{"ignore", ""},
			fetchResp: []string{"argh", "aye"},
			want:      []string{"argh", "aye"},
			wantValue: StringValue(""),
			wantCompleteArgs: map[string]*Value{
				"aight": StringValue(""),
			},
		},
		// Test lists
		{
			name:      "completes all list suggestions for first one",
			args:      []string{"advanced", "liszt", ""},
			fetchResp: []string{"harp", "piano", "picolo"},
			want:      []string{"harp", "piano", "picolo"},
			wantValue: StringListValue(""),
			wantCompleteArgs: map[string]*Value{
				"list-arg": StringListValue(""),
			},
		},
		{
			name:      "completes all list suggestions for later one",
			args:      []string{"advanced", "liszt", "un", "deux", "trois", "quatre", "p"},
			fetchResp: []string{"harp", "piano", "picolo"},
			want:      []string{"piano", "picolo"},
			wantValue: StringListValue("un", "deux", "trois", "quatre", "p"),
			wantCompleteArgs: map[string]*Value{
				"list-arg": StringListValue("un", "deux", "trois", "quatre", "p"),
			},
		},
		// Test extra optional arguments
		{
			name:      "optional argument recommends for minimum",
			args:      []string{"sometimes", ""},
			distinct:  true,
			fetchResp: []string{"occ", "occasional", "temp", "temporary", "tmp"},
			want:      []string{"occ", "occasional", "temp", "temporary", "tmp"},
			wantValue: StringListValue(""),
			wantCompleteArgs: map[string]*Value{
				"opt group": StringListValue(""),
			},
		},
		{
			name:      "optional argument recommends for middle",
			args:      []string{"sometimes", "tmp", "occ", "t"},
			distinct:  true,
			fetchResp: []string{"occ", "occasional", "temp", "temporary", "tmp"},
			want:      []string{"temp", "temporary"},
			wantValue: StringListValue("tmp", "occ", "t"),
			wantCompleteArgs: map[string]*Value{
				"opt group": StringListValue("tmp", "occ", "t"),
			},
		},
		{
			name:      "ignore already listed items",
			args:      []string{"sometimes", "tmp", "occ", "temp", ""},
			distinct:  true,
			fetchResp: []string{"occ", "occasional", "temp", "temporary", "tmp"},
			want:      []string{"occasional", "temporary"},
			wantValue: StringListValue("tmp", "occ", "temp", ""),
			wantCompleteArgs: map[string]*Value{
				"opt group": StringListValue("tmp", "occ", "temp", ""),
			},
		},
		{
			name:      "optional argument recommends for end",
			args:      []string{"sometimes", "tmp", "occ", "temporary", "o"},
			distinct:  true,
			fetchResp: []string{"occ", "occasional", "temp", "temporary", "tmp"},
			want:      []string{"occasional"},
			wantValue: StringListValue("tmp", "occ", "temporary", "o"),
			wantCompleteArgs: map[string]*Value{
				"opt group": StringListValue("tmp", "occ", "temporary", "o"),
			},
		},
		{
			name:      "optional argument does not recommend after limit",
			args:      []string{"sometimes", "tmp", "occ", "temporary", "temp", "extra"},
			distinct:  true,
			fetchResp: []string{"occ", "occasional", "temp", "temporary", "tmp"},
		},
		// Multi-word and quote tests
		{
			name: "multi-word options",
			args: []string{"mw", ""},
			fetchResp: []string{
				"First Choice",
				"Second Thing",
				"Third One",
				"Fourth Option",
				"Fifth",
			},
			want: []string{
				"Fifth",
				`First\ Choice`,
				`Fourth\ Option`,
				`Second\ Thing`,
				`Third\ One`,
			},
			wantValue: StringListValue(""),
			wantCompleteArgs: map[string]*Value{
				"alpha": StringListValue(""),
			},
		},
		{
			name: "last argument matches a multi-word option",
			args: []string{"mw", "Fo"},
			fetchResp: []string{
				"First Choice",
				"Second Thing",
				"Third One",
				"Fourth Option",
				"Fifth",
			},
			want: []string{
				`Fourth\ Option`,
			},
			wantValue: StringListValue("Fo"),
			wantCompleteArgs: map[string]*Value{
				"alpha": StringListValue("Fo"),
			},
		},
		{
			name: "last argument matches several multi-word option",
			args: []string{"mw", "F"},
			fetchResp: []string{
				"First Choice",
				"Second Thing",
				"Third One",
				"Fourth Option",
				"Fifth",
			},
			want: []string{
				"Fifth",
				`First\ Choice`,
				`Fourth\ Option`,
			},
			wantValue: StringListValue("F"),
			wantCompleteArgs: map[string]*Value{
				"alpha": StringListValue("F"),
			},
		},
		{
			name: "args with double quotes count as single option and ignore single quote",
			args: []string{"squo", `"Greg's`, `One"`, ""},
			fetchResp: []string{
				"Greg's One",
				"Greg's Two",
				"Greg's Three",
				"Greg's Four",
			},
			// TODO if string has a quote, then we should escape that as well?
			want: []string{
				`Greg's\ Four`,
				`Greg's\ One`,
				`Greg's\ Three`,
				`Greg's\ Two`,
			},
			wantValue: StringListValue("Greg's One", ""),
			wantCompleteArgs: map[string]*Value{
				"whose": StringListValue("Greg's One", ""),
			},
		},
		{
			name: "args with single quotes count as single option and ignore double quote",
			args: []string{"dquo", `'Greg"s`, `Other"s'`, ""},
			fetchResp: []string{
				`Greg"s One`,
				`Greg"s Two`,
				`Greg"s Three`,
				`Greg"s Four`,
			},
			want: []string{
				`Greg"s\ Four`,
				`Greg"s\ One`,
				`Greg"s\ Three`,
				`Greg"s\ Two`,
			},
			wantValue: StringListValue(`Greg"s Other"s`, ""),
			wantCompleteArgs: map[string]*Value{
				"whose": StringListValue(`Greg"s Other"s`, ""),
			},
		},
		{
			name: "completes properly if ending on double quote",
			args: []string{"mw", `"`},
			fetchResp: []string{
				"First Choice",
				"Second Thing",
				"Third One",
				"Fourth Option",
				"Fifth",
			},
			want: []string{
				"Fifth",
				`"First Choice"`,
				`"Fourth Option"`,
				`"Second Thing"`,
				`"Third One"`,
			},
			wantValue: StringListValue(""),
			wantCompleteArgs: map[string]*Value{
				"alpha": StringListValue(""),
			},
		},
		{
			name: "completes properly if ending on single quote",
			args: []string{"mw", `"First`, `Choice"`, `'`},
			fetchResp: []string{
				"First Choice",
				"Second Thing",
				"Third One",
				"Fourth Option",
				"Fifth",
			},
			want: []string{
				"Fifth",
				"'First Choice'",
				"'Fourth Option'",
				"'Second Thing'",
				"'Third One'",
			},
			wantValue: StringListValue("First Choice", ""),
			wantCompleteArgs: map[string]*Value{
				"alpha": StringListValue("First Choice", ""),
			},
		},
		{
			name: "completes with single quotes if unclosed single quote",
			args: []string{"mw", `"First`, `Choice"`, `'F`},
			fetchResp: []string{
				"First Choice",
				"Second Thing",
				"Third One",
				"Fourth Option",
				"Fifth",
			},
			want: []string{
				"Fifth",
				"'First Choice'",
				"'Fourth Option'",
			},
			wantValue: StringListValue("First Choice", "F"),
			wantCompleteArgs: map[string]*Value{
				"alpha": StringListValue("First Choice", "F"),
			},
		},
		{
			name: "last argument is just a double quote",
			args: []string{"mw", `"`},
			fetchResp: []string{
				"First Choice",
				"Second Thing",
				"Third One",
				"Fourth Option",
				"Fifth",
			},
			want: []string{
				"Fifth",
				`"First Choice"`,
				`"Fourth Option"`,
				`"Second Thing"`,
				`"Third One"`,
			},
			wantValue: StringListValue(""),
			wantCompleteArgs: map[string]*Value{
				"alpha": StringListValue(""),
			},
		},
		{
			name: "last argument is a double quote with words",
			args: []string{"mw", `"F`},
			fetchResp: []string{
				"First Choice",
				"Second Thing",
				"Third One",
				"Fourth Option",
				"Fifth",
			},
			want: []string{
				"Fifth",
				`"First Choice"`,
				`"Fourth Option"`,
			},
			wantValue: StringListValue("F"),
			wantCompleteArgs: map[string]*Value{
				"alpha": StringListValue("F"),
			},
		},
		{
			name: "double quote with single quote",
			args: []string{"squo", `"Greg's T`},
			fetchResp: []string{
				"Greg's One",
				"Greg's Two",
				"Greg's Three",
				"Greg's Four",
			},
			want: []string{
				`"Greg's Three"`,
				`"Greg's Two"`,
			},
			wantValue: StringListValue("Greg's T"),
			wantCompleteArgs: map[string]*Value{
				"whose": StringListValue("Greg's T"),
			},
		},
		{
			name: "last argument is just a single quote",
			args: []string{"mw", "'"},
			fetchResp: []string{
				"First Choice",
				"Second Thing",
				"Third One",
				"Fourth Option",
				"Fifth",
			},
			want: []string{
				"Fifth",
				"'First Choice'",
				"'Fourth Option'",
				"'Second Thing'",
				"'Third One'",
			},
			wantValue: StringListValue(""),
			wantCompleteArgs: map[string]*Value{
				"alpha": StringListValue(""),
			},
		},
		{
			name: "last argument is a single quote with words",
			args: []string{"mw", "'F"},
			fetchResp: []string{
				"First Choice",
				"Second Thing",
				"Third One",
				"Fourth Option",
				"Fifth",
			},
			want: []string{
				"Fifth",
				"'First Choice'",
				"'Fourth Option'",
			},
			wantValue: StringListValue("F"),
			wantCompleteArgs: map[string]*Value{
				"alpha": StringListValue("F"),
			},
		},
		{
			name: "single quote with double quote",
			args: []string{"dquo", `'Greg"s T`},
			fetchResp: []string{
				`Greg"s One`,
				`Greg"s Two`,
				`Greg"s Three`,
				`Greg"s Four`,
			},
			want: []string{
				// TODO: I think this may need backslashes like in the double quote case?
				// test this with actual commands and see what happens
				`'Greg"s Three'`,
				`'Greg"s Two'`,
			},
			wantValue: StringListValue(`Greg"s T`),
			wantCompleteArgs: map[string]*Value{
				"whose": StringListValue(`Greg"s T`),
			},
		},
		{
			name: "end with space",
			args: []string{"prefixes", "Attempt One "},
			fetchResp: []string{
				"Attempt One Two",
				"Attempt OneTwo",
				"Three",
				"Three Four",
				"ThreeFour",
			},
			want: []string{
				`Attempt\ One\ Two`,
			},
			wantValue: StringListValue("Attempt One "),
			wantCompleteArgs: map[string]*Value{
				"alphas": StringListValue("Attempt One "),
			},
		},
		{
			name: "single and double words",
			args: []string{"prefixes", "Three"},
			fetchResp: []string{
				"Attempt One Two",
				"Attempt OneTwo",
				"Three",
				"Three Four",
				"ThreeFour",
			},
			want: []string{
				"Three",
				`Three\ Four`,
				"ThreeFour",
			},
			wantValue: StringListValue("Three"),
			wantCompleteArgs: map[string]*Value{
				"alphas": StringListValue("Three"),
			},
		},
		{
			name: "handles backspaces before spaces",
			args: []string{"mw", "First\\ O"},
			fetchResp: []string{
				"First Of",
				"First One",
				"Second Thing",
				"Third One",
			},
			want: []string{
				`First\ Of`,
				`First\ One`,
			},
			wantValue: StringListValue("First O"),
			wantCompleteArgs: map[string]*Value{
				"alpha": StringListValue("First O"),
			},
		},
		// Flag tests
		{
			name: "completes single hypen with flags",
			args: []string{"intermediate", "-"},
			want: []string{"--american", "--another", "--state"},
		},
		{
			name: "completes double hypen with flags",
			args: []string{"intermediate", "--"},
			want: []string{"--american", "--another", "--state"},
		},
		{
			name: "completes double hypen and prefix with matching flags",
			args: []string{"intermediate", "--a"},
			want: []string{"--american", "--another"},
		},
		{
			name: "completes known short flag",
			args: []string{"intermediate", "-a"},
			want: []string{"-a"},
		},
		{
			name: "flag completion in middle of command",
			args: []string{"intermediate", "hello", "--a"},
			want: []string{"--american", "--another"},
		},
		{
			name:      "regular completion when boolean flag is earlier",
			args:      []string{"intermediate", "--american", "e"},
			fetchResp: []string{"int", "erm", "edi", "ate"},
			want:      []string{"edi", "erm"},
			wantValue: StringListValue("e"),
			wantCompleteArgs: map[string]*Value{
				"syllable": StringListValue("e"),
			},
			wantCompleteFlags: map[string]*Value{
				"american": BoolValue(true),
			},
		},
		{
			name:      "regular completion when short boolean flag is earlier",
			args:      []string{"intermediate", "-a", "e"},
			fetchResp: []string{"int", "erm", "edi", "ate"},
			want:      []string{"edi", "erm"},
			wantValue: StringListValue("e"),
			wantCompleteArgs: map[string]*Value{
				"syllable": StringListValue("e"),
			},
			wantCompleteFlags: map[string]*Value{
				"american": BoolValue(true),
			},
		},
		{
			name:      "regular completion when flag with argument is earlier",
			args:      []string{"intermediate", "ate", "--state", "maine", "e"},
			fetchResp: []string{"int", "erm", "edi", "ate"},
			want:      []string{"edi", "erm"},
			wantValue: StringListValue("ate", "e"),
			wantCompleteArgs: map[string]*Value{
				"syllable": StringListValue("ate", "e"),
			},
			wantCompleteFlags: map[string]*Value{
				"state": StringListValue("maine"),
			},
		},
		{
			name:      "regular completion when short flag with argument is earlier",
			args:      []string{"intermediate", "ate", "-s", "maine", "e"},
			fetchResp: []string{"int", "erm", "edi", "ate"},
			want:      []string{"edi", "erm"},
			wantValue: StringListValue("ate", "e"),
			wantCompleteArgs: map[string]*Value{
				"syllable": StringListValue("ate", "e"),
			},
			wantCompleteFlags: map[string]*Value{
				"state": StringListValue("maine"),
			},
		},
		{
			name:      "flag arguments are autocompleted",
			args:      []string{"intermediate", "--state", ""},
			fetchResp: []string{"california", "connecticut", "washington", "washington_dc"},
			want:      []string{"california", "connecticut", "washington", "washington_dc"},
			wantValue: StringListValue(""),
			wantCompleteFlags: map[string]*Value{
				"state": StringListValue(""),
			},
		},
		{
			name:      "partial flag arguments are autocompleted",
			args:      []string{"intermediate", "--state", "c"},
			fetchResp: []string{"california", "connecticut"},
			want:      []string{"california", "connecticut"},
			wantValue: StringListValue("c"),
			wantCompleteFlags: map[string]*Value{
				"state": StringListValue("c"),
			},
		},
		{
			name:      "short flag arguments are autocompleted",
			args:      []string{"intermediate", "-s", "washington"},
			fetchResp: []string{"california", "connecticut", "washington", "washington_dc"},
			want:      []string{"washington", "washington_dc"},
			wantValue: StringListValue("washington"),
			wantCompleteFlags: map[string]*Value{
				"state": StringListValue("washington"),
			},
		},
		{
			name:      "flag completion works when several flags",
			args:      []string{"intermediate", "--another", "a", "int", "erm", "-a", "edi", "--state", ""},
			fetchResp: []string{"california", "connecticut", "washington", "washington_dc"},
			want:      []string{"california", "connecticut", "washington", "washington_dc"},
			wantValue: StringListValue(""),
			wantCompleteFlags: map[string]*Value{
				"american": BoolValue(true),
				"another":  StringListValue("a"),
				"state":    StringListValue(""),
			},
		},
		{
			name:      "flag completion works when several flags and partial flag arg",
			args:      []string{"intermediate", "--another", "a", "int", "erm", "-a", "edi", "--state", "wash"},
			fetchResp: []string{"california", "connecticut", "washington", "washington_dc"},
			want:      []string{"washington", "washington_dc"},
			wantValue: StringListValue("wash"),
			wantCompleteFlags: map[string]*Value{
				"american": BoolValue(true),
				"another":  StringListValue("a"),
				"state":    StringListValue("wash"),
			},
		},
		{
			name:      "flag completion works when flag has multiple arguments",
			args:      []string{"wave", "--yourFlag", ""},
			fetchResp: []string{"please", "person", "okay"},
			want:      []string{"okay", "person", "please"},
			wantValue: StringListValue(""),
			wantCompleteFlags: map[string]*Value{
				"yourFlag": StringListValue(""),
			},
		},
		{
			name:      "flag partial completion works when flag has multiple arguments",
			args:      []string{"wave", "--yourFlag", "p"},
			fetchResp: []string{"please", "person", "okay"},
			want:      []string{"person", "please"},
			wantValue: StringListValue("p"),
			wantCompleteFlags: map[string]*Value{
				"yourFlag": StringListValue("p"),
			},
		},
		// BranchCommand tests
		{
			name:      "autocompletes nested branch command",
			args:      []string{"advanced", ""},
			fetchResp: []string{"forVoting", "forDriving", "somethingElse"},
			want: []string{
				"first",
				"forDriving",
				"forVoting",
				"foremost",
				"liszt",
				"other",
				"somethingElse",
			},
			wantCompleteArgs: map[string]*Value{
				"cb-command": StringListValue(""),
			},
			wantValue: StringListValue(""),
		},
		{
			name:      "autocompletes nested branch command with partial completiong",
			args:      []string{"advanced", "f"},
			fetchResp: []string{"forVoting", "forDriving", "somethingElse"},
			want: []string{
				"first",
				"forDriving",
				"forVoting",
				"foremost",
			},
			wantCompleteArgs: map[string]*Value{
				"cb-command": StringListValue("f"),
			},
			wantValue: StringListValue("f"),
		},
		{
			name:      "autocompletes only terminus command if no subcommand match",
			args:      []string{"advanced", "noMatch", ""},
			fetchResp: []string{"forVoting", "forDriving", "somethingElse"},
			want: []string{
				"forDriving",
				"forVoting",
				"somethingElse",
			},
			wantCompleteArgs: map[string]*Value{
				"cb-command": StringListValue("noMatch", ""),
			},
			wantValue: StringListValue("noMatch", ""),
		},
		{
			name:      "autocompletes partial only terminus command if no subcommand match",
			args:      []string{"advanced", "noMatch", "for"},
			fetchResp: []string{"forVoting", "forDriving", "somethingElse"},
			want: []string{
				"forDriving",
				"forVoting",
			},
			wantCompleteArgs: map[string]*Value{
				"cb-command": StringListValue("noMatch", "for"),
			},
			wantValue: StringListValue("noMatch", "for"),
		},
		{
			name: "autocompletes handles invalid branch command",
			args: []string{"not", "a", "key"},
		},
		// Commands with different value type.
		// string argument type
		{
			name:      "completes string argument",
			args:      []string{"valueTypes", "string", ""},
			fetchResp: []string{"hi", "hello"},
			want:      []string{"hello", "hi"},
			wantValue: StringValue(""),
			wantCompleteArgs: map[string]*Value{
				"req": StringValue(""),
			},
		},
		{
			name:      "completes optional string argument",
			args:      []string{"valueTypes", "string", "hello", ""},
			fetchResp: []string{"world", "there", "toYou"},
			want:      []string{"there", "toYou", "world"},
			wantValue: StringValue(""),
			wantCompleteArgs: map[string]*Value{
				"req": StringValue("hello"),
				"opt": StringValue(""),
			},
		},
		// stringList argument type
		{
			name:      "completes stringList argument",
			args:      []string{"valueTypes", "stringList", "hello", ""},
			fetchResp: []string{"there", "world"},
			want:      []string{"there", "world"},
			wantValue: StringListValue("hello", ""),
			wantCompleteArgs: map[string]*Value{
				"req": StringListValue("hello", ""),
			},
		},
		{
			name:      "completes optional stringList argument",
			args:      []string{"valueTypes", "stringList", "hello", "to", ""},
			fetchResp: []string{"them", "you"},
			want:      []string{"them", "you"},
			wantValue: StringListValue("hello", "to", ""),
			wantCompleteArgs: map[string]*Value{
				"req": StringListValue("hello", "to", ""),
			},
		},
		// int argument type
		{
			name:      "completes int argument",
			args:      []string{"valueTypes", "int", ""},
			fetchResp: []string{"123", "456"},
			want:      []string{"123", "456"},
			wantValue: IntValue(0),
			wantCompleteArgs: map[string]*Value{
				"req": IntValue(0),
			},
		},
		{
			name:      "completes optional int argument",
			args:      []string{"valueTypes", "int", "123", ""},
			fetchResp: []string{"45", "678"},
			want:      []string{"45", "678"},
			wantValue: IntValue(0),
			wantCompleteArgs: map[string]*Value{
				"req": IntValue(123),
				"opt": IntValue(0),
			},
		},
		{
			name:      "completes when previous int was bad format",
			args:      []string{"valueTypes", "int", "123.45", ""},
			fetchResp: []string{"45", "678"},
			want:      []string{"45", "678"},
			wantValue: IntValue(0),
			wantCompleteArgs: map[string]*Value{
				"req": IntValue(0),
				"opt": IntValue(0),
			},
		},
		// intList argument type
		{
			name:      "completes intList argument",
			args:      []string{"valueTypes", "intList", "123", ""},
			fetchResp: []string{"45", "678"},
			want:      []string{"45", "678"},
			wantValue: IntListValue(123, 0),
			wantCompleteArgs: map[string]*Value{
				"req": IntListValue(123, 0),
			},
		},
		{
			name:      "completes optional intList argument",
			args:      []string{"valueTypes", "intList", "123", "45", ""},
			fetchResp: []string{"67", "89"},
			want:      []string{"67", "89"},
			wantValue: IntListValue(123, 45, 0),
			wantCompleteArgs: map[string]*Value{
				"req": IntListValue(123, 45, 0),
			},
		},
		{
			name:      "completes intList when previous argument is invalid",
			args:      []string{"valueTypes", "intList", "twelve", "45", ""},
			fetchResp: []string{"67", "89"},
			want:      []string{"67", "89"},
			wantValue: IntListValue(0, 45, 0),
			wantCompleteArgs: map[string]*Value{
				"req": IntListValue(0, 45, 0),
			},
		},
		// float argument type
		{
			name:      "completes float argument",
			args:      []string{"valueTypes", "float", ""},
			fetchResp: []string{"12.3", "-456"},
			want:      []string{"-456", "12.3"},
			wantValue: FloatValue(0),
			wantCompleteArgs: map[string]*Value{
				"req": FloatValue(0),
			},
		},
		{
			name:      "completes optional float argument",
			args:      []string{"valueTypes", "float", "1.23", ""},
			fetchResp: []string{"-4.5", "678"},
			want:      []string{"-4.5", "678"},
			wantValue: FloatValue(0),
			wantCompleteArgs: map[string]*Value{
				"req": FloatValue(1.23),
				"opt": FloatValue(0),
			},
		},
		{
			name:      "completes when previous float was bad format",
			args:      []string{"valueTypes", "float", "eleven", ""},
			fetchResp: []string{"-45", "67.8"},
			want:      []string{"-45", "67.8"},
			wantValue: FloatValue(0),
			wantCompleteArgs: map[string]*Value{
				"req": FloatValue(0),
				"opt": FloatValue(0),
			},
		},
		// floatList argument type
		{
			name:      "completes floatList argument",
			args:      []string{"valueTypes", "floatList", "123.", ""},
			fetchResp: []string{"4.5", "-678."},
			want:      []string{"-678.", "4.5"},
			wantValue: FloatListValue(123, 0),
			wantCompleteArgs: map[string]*Value{
				"req": FloatListValue(123, 0),
			},
		},
		{
			name:      "completes optional floatList argument",
			args:      []string{"valueTypes", "floatList", "0.123", "-.45", ""},
			fetchResp: []string{".67", "-.89"},
			want:      []string{"-.89", ".67"},
			wantValue: FloatListValue(0.123, -0.45, 0),
			wantCompleteArgs: map[string]*Value{
				"req": FloatListValue(0.123, -0.45, 0),
			},
		},
		{
			name:      "completes floatList when previous argument is invalid",
			args:      []string{"valueTypes", "floatList", "twelve", "6.7", ""},
			fetchResp: []string{"6.7", "89"},
			distinct:  true,
			want:      []string{"6.7", "89"},
			wantValue: FloatListValue(0, 6.7, 0),
			wantCompleteArgs: map[string]*Value{
				"req": FloatListValue(0, 6.7, 0),
			},
		},
		// bool argument type
		{
			name: "completes bool argument",
			args: []string{"valueTypes", "bool", ""},
			want: []string{"f", "false", "t", "true"},
		},
		{
			name: "completes partial bool argument",
			args: []string{"valueTypes", "bool", "t"},
			want: []string{"t", "true"},
		},
		{
			name: "completes optional boolean argument",
			args: []string{"valueTypes", "bool", "true", "f"},
			want: []string{"f", "false"},
			// TODO: this doesn't test wantCompleteArgs, because completor is builtin.
		},
		{
			name: "completes when previous boolean was bad format",
			args: []string{"valueTypes", "bool", "maybe", ""},
			want: []string{"f", "false", "t", "true"},
		},
		/* Useful comment for commenting out tests */
	} {
		t.Run(test.name, func(t *testing.T) {
			fetcher := &testFetcher{
				resp: test.fetchResp,
			}
			completor := &Completor{
				Distinct:          test.distinct,
				SuggestionFetcher: fetcher,
			}

			got := Autocomplete(branchCommand(NoopExecutor, completor), test.args, test.cursorIdx)
			if len(got) == 0 {
				got = nil
			}
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("command.Autocomplete(%v, %d) returned diff (-want, +got):\n%s", test.args, test.cursorIdx, diff)
			}

			if diff := cmp.Diff(test.wantCompleteArgs, fetcher.gotArgs); diff != "" {
				t.Errorf("command.Autocomplete(%v, %d) produced complete args diff (-want +got):\n%s", test.args, test.cursorIdx, diff)
			}

			if diff := cmp.Diff(test.wantCompleteFlags, fetcher.gotFlags); diff != "" {
				t.Errorf("command.Autocomplete(%v, %d) produced complete flags diff (-want +got):\n%s", test.args, test.cursorIdx, diff)
			}

			if diff := cmp.Diff(test.wantValue, fetcher.gotValue); diff != "" {
				t.Errorf("command.Autocomplete(%v, %d) produced values diff (-want +got):\n%s", test.args, test.cursorIdx, diff)
			}
		})
	}
}

type testFetcher struct {
	gotValue *Value
	gotArgs  map[string]*Value
	gotFlags map[string]*Value
	resp     []string
}

func (tf *testFetcher) Fetch(value *Value, args, flags map[string]*Value) *Completion {
	// Check length so we can consider empty to be the same as nil.
	// That makes for cleaner test cases.
	if value != nil && value.Length() > 0 {
		tf.gotValue = value
	}
	if len(args) > 0 {
		tf.gotArgs = args
	}
	if len(flags) > 0 {
		tf.gotFlags = flags
	}

	return &Completion{
		Suggestions: tf.resp,
	}
}

// Test to get 100% coverage
func TestMiscellaneous(t *testing.T) {
	t.Run("NoopExecutor returns nothing", func(t *testing.T) {
		args := map[string]*Value{
			"a": StringListValue("b"),
			"c": StringListValue("d", "e"),
		}
		flags := map[string]*Value{
			"f0": StringListValue(),
			"f1": StringListValue("12", "3"),
			"f4": StringListValue("4", "56"),
		}

		if resp, ok := NoopExecutor(nil, args, flags, nil); resp != nil && !ok {
			t.Errorf("Expected NoopExecutor to return (nil, true); got (%v, %v)", resp, ok)
		}
	})

	t.Run("completor with nil fetch options", func(t *testing.T) {
		c := &Completor{}
		v := StringListValue("he", "yo")
		as := map[string]*Value{
			"hey": StringListValue("oooo"),
		}
		fs := map[string]*Value{
			"hey": StringListValue("o", "o"),
		}
		_ = c.Complete("yo", v, as, fs)
	})
}
