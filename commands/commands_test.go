package commands

// TODO: split this up into separate files (not separate packages).

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestUsage(t *testing.T) {
	for _, test := range []struct {
		name string
		cmd  Command
		want []string
	}{
		// Basic tests
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
				"basic", "POS_1", "PLACE_2", "--american|-a", "--another", "FLAG_VALUE", "--state|-s", "FLAG_VALUE", "\n",
				"basically", "ANYTHING", "ANYTHING", "ANYTHING", "\n",
				"beginner", "\n", "dquo", "WHOSE", "WHOSE", "\n",
				"ignore", "alpha", "\n", "ayo", "\n", "AIGHT", "\n",
				"intermediate", "SYLLABLE", "SYLLABLE", "SYLLABLE", "--american|-a", "--another", "FLAG_VALUE", "--state|-s", "FLAG_VALUE", "\n",
				"mw", "ALPHA", "ALPHA", "\n",
				"prefixes", "ALPHAS", "\n",
				"sometimes", "OPT_GROUP", "[", "OPT_GROUP", "OPT_GROUP", "OPT_GROUP", "]", "\n",
				"squo", "WHOSE", "WHOSE", "\n",
				"valueTypes",
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
							// TODO: make -1 a constant for unbounded list.
							StringListArg("list-arg", 1, -1, completor, opts...),
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
					// TODO: make these file names. Current values are more like BranchOmmands.
					StringListArg("pos_1", 1, 0, completor, opts...),
					StringListArg("place 2", 1, 0, completor, opts...),
				},
				Flags: []Flag{
					NewBooleanFlag("american", 'a', false),
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
					NewBooleanFlag("american", 'a', false),
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
				},
			},
		},
	}
}

// TODO: combine execute and autocomplete tests?
func TestExecute(t *testing.T) {

	for _, test := range []struct {
		name             string
		args             []string
		ex               Executor
		exResp           *ExecutorResponse
		exErr            error
		opts             []ArgOpt
		want             *ExecutorResponse
		wantExecuteArgs  map[string]*Value
		wantExecuteFlags map[string]*Value
		wantErr          string
	}{
		// Basic tests
		{
			name:    "empty args",
			wantErr: "more args required",
		},
		{
			name:    "incomplete command",
			args:    []string{"huh"},
			wantErr: `unknown subcommand and no terminus command defined`,
		},
		{
			name:    "not enough flag values",
			args:    []string{"basic", "--state"},
			wantErr: `not enough values passed to flag "state"`,
		},
		{
			name:    "too many positional arguments",
			args:    []string{"basic", "--state", "maine", "build", "one", "else", "too"},
			wantErr: "extra unknown args ([else too])",
		},
		{
			name:    "not enough positional arguments",
			args:    []string{"intermediate", "--state", "maine", "one"},
			wantErr: `not enough arguments for "syllable" arg`,
		},
		{
			name:    "not enough positional arguments",
			args:    []string{"basic", "--state", "maine"},
			wantErr: `no argument provided for "pos_1"`,
		},
		{
			name:    "no executor defined",
			args:    []string{"advanced", "other"},
			wantErr: "no executor defined for command",
		},
		{
			name: "works when CommandBranch defines terminusCommand",
			args: []string{"advanced", "not", "registered"},
			wantExecuteArgs: map[string]*Value{
				"cb-command": &Value{
					stringList: []string{"not", "registered"},
				},
			},
		},
		{
			name: "fails when CommandBranch defines executor fails",
			ex: func(args map[string]*Value, flags map[string]*Value) (*ExecutorResponse, error) {
				return nil, fmt.Errorf("bad news bears")
			},
			args:    []string{"advanced", "not", "registered"},
			wantErr: "bad news bears",
		},
		{
			name: "works with no flags",
			args: []string{"basic", "un", "deux"},
			wantExecuteArgs: map[string]*Value{
				"pos_1": &Value{
					stringList: []string{"un"},
				},
				"place 2": &Value{
					stringList: []string{"deux"},
				},
			},
		},
		{
			name: "works with flags at the beginning",
			args: []string{"basic", "--state", "jersey", "trois", "quatre"},
			wantExecuteArgs: map[string]*Value{
				"pos_1": &Value{
					stringList: []string{"trois"},
				},
				"place 2": &Value{
					stringList: []string{"quatre"},
				},
			},
			wantExecuteFlags: map[string]*Value{
				"state": &Value{
					stringList: []string{"jersey"},
				},
			},
		},
		{
			name: "works with flags in the middle",
			args: []string{"basic", "trois", "--state", "massachusetts", "quatre"},
			wantExecuteArgs: map[string]*Value{
				"pos_1": &Value{
					stringList: []string{"trois"},
				},
				"place 2": &Value{
					stringList: []string{"quatre"},
				},
			},
			wantExecuteFlags: map[string]*Value{
				"state": &Value{
					stringList: []string{"massachusetts"},
				},
			},
		},
		{
			name: "works with flags at the end",
			args: []string{"basic", "trois", "quatre", "-s", "connecticut"},
			wantExecuteArgs: map[string]*Value{
				"pos_1": &Value{
					stringList: []string{"trois"},
				},
				"place 2": &Value{
					stringList: []string{"quatre"},
				},
			},
			wantExecuteFlags: map[string]*Value{
				"state": &Value{
					stringList: []string{"connecticut"},
				},
			},
		},
		{
			name: "works with boolean flag",
			args: []string{"basic", "trois", "--american", "quatre"},
			wantExecuteArgs: map[string]*Value{
				"pos_1": &Value{
					stringList: []string{"trois"},
				},
				"place 2": &Value{
					stringList: []string{"quatre"},
				},
			},
			wantExecuteFlags: map[string]*Value{
				"american": &Value{
					stringList: []string{},
				},
			},
		},
		{
			name: "works with short boolean flag",
			args: []string{"basic", "-a", "trois", "quatre"},
			wantExecuteArgs: map[string]*Value{
				"pos_1": &Value{
					stringList: []string{"trois"},
				},
				"place 2": &Value{
					stringList: []string{"quatre"},
				},
			},
			wantExecuteFlags: map[string]*Value{
				"american": &Value{
					stringList: []string{},
				},
			},
		},
		{
			name: "works with arguments with multiple args",
			args: []string{"intermediate", "first", "2nd", "bronze"},
			wantExecuteArgs: map[string]*Value{
				"syllable": &Value{
					stringList: []string{"first", "2nd", "bronze"},
				},
			},
		},
		// Test lists
		{
			name:    "list fails when not enough args",
			args:    []string{"advanced", "liszt"},
			wantErr: `no argument provided for "list-arg"`,
		},
		{
			name: "list succeeds when at minimum args",
			args: []string{"advanced", "liszt", "piano"},
			wantExecuteArgs: map[string]*Value{
				"list-arg": &Value{
					stringList: []string{"piano"},
				},
			},
		},
		{
			name: "list succeeds when extra args",
			args: []string{"advanced", "liszt", "piano", "harp", "picolo"},
			wantExecuteArgs: map[string]*Value{
				"list-arg": &Value{
					stringList: []string{"piano", "harp", "picolo"},
				},
			},
		},
		{
			name: "list succeeds when flag in between",
			args: []string{"advanced", "liszt", "piano", "--inside", "56", "34", "harp", "picolo"},
			wantExecuteArgs: map[string]*Value{
				"list-arg": &Value{
					stringList: []string{"piano", "harp", "picolo"},
				},
			},
			wantExecuteFlags: map[string]*Value{
				"inside": &Value{
					stringList: []string{"56", "34"},
				},
			},
		},
		{
			name: "list succeeds when short flag in between",
			args: []string{"advanced", "liszt", "piano", "-i", "56", "34", "harp", "picolo"},
			wantExecuteArgs: map[string]*Value{
				"list-arg": &Value{
					stringList: []string{"piano", "harp", "picolo"},
				},
			},
			wantExecuteFlags: map[string]*Value{
				"inside": &Value{
					stringList: []string{"56", "34"},
				},
			},
		},
		// Test extra optional arguments.
		{
			name:    "optional argument doesn't accept less than minimum",
			args:    []string{"sometimes"},
			wantErr: `no argument provided for "opt group"`,
		},
		{
			name: "optional argument accepts minimum",
			args: []string{"sometimes", "temp"},
			wantExecuteArgs: map[string]*Value{
				"opt group": &Value{
					stringList: []string{"temp"},
				},
			},
		},
		{
			name: "optional argument accepts middle amount",
			args: []string{"sometimes", "temp", "occasional"},
			wantExecuteArgs: map[string]*Value{
				"opt group": &Value{
					stringList: []string{"temp", "occasional"},
				},
			},
		},
		{
			name: "optional argument accepts max amount",
			args: []string{"sometimes", "temp", "occasional", "tmp", "temporary"},
			wantExecuteArgs: map[string]*Value{
				"opt group": &Value{
					stringList: []string{"temp", "occasional", "tmp", "temporary"},
				},
			},
		},
		{
			name:    "optional argument does not accept more than max amount",
			args:    []string{"sometimes", "temp", "occasional", "tmp", "temporary", "occ"},
			wantErr: "extra unknown args ([occ])",
		},
		// Test return values
		{
			name: "returns what the executor returns",
			args: []string{"intermediate", "first", "2nd", "bronze"},
			wantExecuteArgs: map[string]*Value{
				"syllable": &Value{
					stringList: []string{"first", "2nd", "bronze"},
				},
			},
			exResp: &ExecutorResponse{Stdout: []string{"this", "was a", "success"}},
			want:   &ExecutorResponse{Stdout: []string{"this", "was a", "success"}},
		},
		{
			name: "fails when executor returns an error",
			args: []string{"intermediate", "first", "2nd", "bronze"},
			wantExecuteArgs: map[string]*Value{
				"syllable": &Value{
					stringList: []string{"first", "2nd", "bronze"},
				},
			},
			exErr:   fmt.Errorf("this was a failure"),
			wantErr: "this was a failure",
		},
		// CommandBranch with terminus command
		{
			name: "branch command's terminus command with no arguments",
			args: []string{"advanced"},
		},
		// Commands with different value types.
		// string argument type
		{
			name: "handles string argument",
			args: []string{"valueTypes", "string", "hello"},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType:   StringType,
					stringVal: "hello",
				},
			},
		},
		{
			name: "handles optional string argument",
			args: []string{"valueTypes", "string", "hello", "there"},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType:   StringType,
					stringVal: "hello",
				},
				"opt": &Value{
					valType:   StringType,
					stringVal: "there",
				},
			},
		},
		// stringList argument type
		{
			name: "handles stringList argument",
			args: []string{"valueTypes", "stringList", "its", "me"},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType:    StringListType,
					stringList: []string{"its", "me"},
				},
			},
		},
		{
			name: "handles optional stringList arguments",
			args: []string{"valueTypes", "stringList", "its", "me", "mario"},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType:    StringListType,
					stringList: []string{"its", "me", "mario"},
				},
			},
		},
		// int argument type
		{
			name: "handles int argument",
			args: []string{"valueTypes", "int", "123"},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType: IntType,
					intVal:  123,
				},
			},
		},
		{
			name: "handles optional int argument",
			args: []string{"valueTypes", "int", "123", "-45"},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType: IntType,
					intVal:  123,
				},
				"opt": &Value{
					valType: IntType,
					intVal:  -45,
				},
			},
		},
		{
			name:    "int argument requires int value",
			args:    []string{"valueTypes", "int", "123.45"},
			wantErr: `failed to process args: failed to convert value: argument should be an integer: strconv.Atoi: parsing "123.45": invalid syntax`,
		},
		{
			name:    "int flag requires int value",
			args:    []string{"valueTypes", "int", "-v", "123.45"},
			wantErr: `failed to process flags: failed to convert value: argument should be an integer: strconv.Atoi: parsing "123.45": invalid syntax`,
		},
		// intList argument type
		{
			name: "handles intList argument",
			args: []string{"valueTypes", "intList", "123", "-45"},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType: IntListType,
					intList: []int{123, -45},
				},
			},
		},
		{
			name: "handles optional intList arguments",
			args: []string{"valueTypes", "intList", "123", "-45", "0"},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType: IntListType,
					intList: []int{123, -45, 0},
				},
			},
		},
		{
			name:    "int list argument requires int values",
			args:    []string{"valueTypes", "intList", "-10", "123.45"},
			wantErr: `failed to process args: failed to convert value: int required for IntList argument type: strconv.Atoi: parsing "123.45": invalid syntax`,
		},
		{
			name:    "int list argument requires int values",
			args:    []string{"valueTypes", "intList", "-v", "123.45"},
			wantErr: `failed to process flags: failed to convert value: int required for IntList argument type: strconv.Atoi: parsing "123.45": invalid syntax`,
		},
		// float argument type
		{
			name: "handles float argument",
			args: []string{"valueTypes", "float", "123.45"},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType:  FloatType,
					floatVal: 123.45,
				},
			},
		},
		{
			name: "handles optional float argument",
			args: []string{"valueTypes", "float", "123.45", "-67"},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType:  FloatType,
					floatVal: 123.45,
				},
				"opt": &Value{
					valType:  FloatType,
					floatVal: -67,
				},
			},
		},
		{
			name:    "float argument requires float value",
			args:    []string{"valueTypes", "float", "twelve"},
			wantErr: `failed to process args: failed to convert value: argument should be a float: strconv.ParseFloat: parsing "twelve": invalid syntax`,
		},
		{
			name:    "float flag requires float value",
			args:    []string{"valueTypes", "float", "--vFlag", "twelve"},
			wantErr: `failed to process flags: failed to convert value: argument should be a float: strconv.ParseFloat: parsing "twelve": invalid syntax`,
		},
		// floatList argument type
		{
			name: "handles floatList argument",
			args: []string{"valueTypes", "floatList", "123.45", "-67"},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType:   FloatListType,
					floatList: []float64{123.45, -67},
				},
			},
		},
		{
			name: "handles optional floatList arguments",
			args: []string{"valueTypes", "floatList", "123.45", "-67", "0"},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType:   FloatListType,
					floatList: []float64{123.45, -67, 0},
				},
			},
		},
		{
			name:    "float list argument requires float values",
			args:    []string{"valueTypes", "floatList", "-10", "twelve"},
			wantErr: `failed to process args: failed to convert value: float required for FloatList argument type: strconv.ParseFloat: parsing "twelve": invalid syntax`,
		},
		{
			name:    "float list flag requires float values",
			args:    []string{"valueTypes", "floatList", "-v", "twelve"},
			wantErr: `failed to process flags: failed to convert value: float required for FloatList argument type: strconv.ParseFloat: parsing "twelve": invalid syntax`,
		},
		// ArgOpt tests
		{
			name: "Breaks when arg option is for invalid type",
			args: []string{"valueTypes", "string", "123"},
			opts: []ArgOpt{
				IntEQ(123),
			},
			wantErr: "failed to process args: failed to convert value: option can only be bound to arguments with type 2",
		},
		// Contains
		{
			name: "Contains works",
			args: []string{"valueTypes", "string", "goodbye"},
			opts: []ArgOpt{
				Contains("good"),
			},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType:   StringType,
					stringVal: "goodbye",
				},
			},
		},
		{
			name: "Contains fails",
			args: []string{"valueTypes", "string", "hello"},
			opts: []ArgOpt{
				Contains("good"),
			},
			wantErr: `failed to process args: failed to convert value: validation failed: [Contains] value doesn't contain substring "good"`,
		},
		// MinLength
		{
			name: "MinLength fails if too few characters",
			args: []string{"valueTypes", "string", "ab"},
			opts: []ArgOpt{
				MinLength(3),
			},
			wantErr: `failed to process args: failed to convert value: validation failed: [MinLength] value must be at least 3 characters`,
		},
		{
			name: "MinLength passes when exact number of characters",
			args: []string{"valueTypes", "string", "abc"},
			opts: []ArgOpt{
				MinLength(3),
			},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType:   StringType,
					stringVal: "abc",
				},
			},
		},
		{
			name: "MinLength passes when exact number of characters",
			args: []string{"valueTypes", "string", "abcd"},
			opts: []ArgOpt{
				MinLength(3),
			},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType:   StringType,
					stringVal: "abcd",
				},
			},
		},
		// IntEQ
		{
			name: "IntEQ works",
			args: []string{"valueTypes", "int", "24"},
			opts: []ArgOpt{
				IntEQ(24),
			},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType: IntType,
					intVal:  24,
				},
			},
		},
		{
			name: "IntEQ fails",
			args: []string{"valueTypes", "int", "25"},
			opts: []ArgOpt{
				IntEQ(24),
			},
			wantErr: "failed to process args: failed to convert value: validation failed: [IntEQ] value isn't equal to 24",
		},
		// IntNE
		{
			name: "IntNE works",
			args: []string{"valueTypes", "int", "24"},
			opts: []ArgOpt{
				IntNE(25),
			},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType: IntType,
					intVal:  24,
				},
			},
		},
		{
			name: "IntNE fails",
			args: []string{"valueTypes", "int", "25"},
			opts: []ArgOpt{
				IntNE(25),
			},
			wantErr: "failed to process args: failed to convert value: validation failed: [IntNE] value isn't not equal to 25",
		},
		// IntLT
		{
			name: "IntLT works",
			args: []string{"valueTypes", "int", "24"},
			opts: []ArgOpt{
				IntLT(25),
			},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType: IntType,
					intVal:  24,
				},
			},
		},
		{
			name: "IntLT fails when equal",
			args: []string{"valueTypes", "int", "25"},
			opts: []ArgOpt{
				IntLT(25),
			},
			wantErr: "failed to process args: failed to convert value: validation failed: [IntLT] value isn't less than 25",
		},
		{
			name: "IntLT fails when not less",
			args: []string{"valueTypes", "int", "26"},
			opts: []ArgOpt{
				IntLT(25),
			},
			wantErr: "failed to process args: failed to convert value: validation failed: [IntLT] value isn't less than 25",
		},
		// IntLTE
		{
			name: "IntLTE works",
			args: []string{"valueTypes", "int", "24"},
			opts: []ArgOpt{
				IntLTE(25),
			},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType: IntType,
					intVal:  24,
				},
			},
		},
		{
			name: "IntLTE works when equal",
			args: []string{"valueTypes", "int", "25"},
			opts: []ArgOpt{
				IntLTE(25),
			},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType: IntType,
					intVal:  25,
				},
			},
		},
		{
			name: "IntLT fails when not less",
			args: []string{"valueTypes", "int", "26"},
			opts: []ArgOpt{
				IntLTE(25),
			},
			wantErr: "failed to process args: failed to convert value: validation failed: [IntLTE] value isn't less than or equal to 25",
		},
		// IntLT
		{
			name: "IntGT fails when not greater",
			args: []string{"valueTypes", "int", "24"},
			opts: []ArgOpt{
				IntGT(25),
			},
			wantErr: "failed to process args: failed to convert value: validation failed: [IntGT] value isn't greater than 25",
		},
		{
			name: "IntGT fails when equal",
			args: []string{"valueTypes", "int", "25"},
			opts: []ArgOpt{
				IntGT(25),
			},
			wantErr: "failed to process args: failed to convert value: validation failed: [IntGT] value isn't greater than 25",
		},
		{
			name: "IntGT works",
			args: []string{"valueTypes", "int", "26"},
			opts: []ArgOpt{
				IntGT(25),
			},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType: IntType,
					intVal:  26,
				},
			},
		},
		// IntGTE
		{
			name: "IntGTE fails when not greater",
			args: []string{"valueTypes", "int", "24"},
			opts: []ArgOpt{
				IntGTE(25),
			},
			wantErr: "failed to process args: failed to convert value: validation failed: [IntGTE] value isn't greater than or equal to 25",
		},
		{
			name: "IntGTE works when equal",
			args: []string{"valueTypes", "int", "25"},
			opts: []ArgOpt{
				IntGTE(25),
			},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType: IntType,
					intVal:  25,
				},
			},
		},
		{
			name: "IntGTE works",
			args: []string{"valueTypes", "int", "26"},
			opts: []ArgOpt{
				IntGTE(25),
			},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType: IntType,
					intVal:  26,
				},
			},
		},
		// IntPositive
		{
			name: "IntPositive fails when negative",
			args: []string{"valueTypes", "int", "-1"},
			opts: []ArgOpt{
				IntPositive(),
			},
			wantErr: "failed to process args: failed to convert value: validation failed: [IntPositive] value isn't positive",
		},
		{
			name: "IntPositive fails when zero",
			args: []string{"valueTypes", "int", "0"},
			opts: []ArgOpt{
				IntPositive(),
			},
			wantErr: "failed to process args: failed to convert value: validation failed: [IntPositive] value isn't positive",
		},
		{
			name: "IntPositive works when positive",
			args: []string{"valueTypes", "int", "1"},
			opts: []ArgOpt{
				IntPositive(),
			},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType: IntType,
					intVal:  1,
				},
			},
		},
		// IntNegative
		{
			name: "IntNegative works when negative",
			args: []string{"valueTypes", "int", "-1"},
			opts: []ArgOpt{
				IntNegative(),
			},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType: IntType,
					intVal:  -1,
				},
			},
		},
		{
			name: "IntNegative fails when zero",
			args: []string{"valueTypes", "int", "0"},
			opts: []ArgOpt{
				IntNegative(),
			},
			wantErr: "failed to process args: failed to convert value: validation failed: [IntNegative] value isn't negative",
		},
		{
			name: "IntNegative fails when positive",
			args: []string{"valueTypes", "int", "1"},
			opts: []ArgOpt{
				IntNegative(),
			},
			wantErr: "failed to process args: failed to convert value: validation failed: [IntNegative] value isn't negative",
		},
		// IntNonNegative
		{
			name: "IntNonNegative fails when negative",
			args: []string{"valueTypes", "int", "-1"},
			opts: []ArgOpt{
				IntNonNegative(),
			},
			wantErr: "failed to process args: failed to convert value: validation failed: [IntNonNegative] value isn't non-negative",
		},
		{
			name: "IntNonNegative works when zero",
			args: []string{"valueTypes", "int", "0"},
			opts: []ArgOpt{
				IntNonNegative(),
			},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType: IntType,
					intVal:  0,
				},
			},
		},
		{
			name: "IntNonNegative works when positive",
			args: []string{"valueTypes", "int", "1"},
			opts: []ArgOpt{
				IntNonNegative(),
			},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType: IntType,
					intVal:  1,
				},
			},
		},
		// FloatEQ
		{
			name: "FloatEQ works",
			args: []string{"valueTypes", "float", "24"},
			opts: []ArgOpt{
				FloatEQ(24),
			},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType:  FloatType,
					floatVal: 24,
				},
			},
		},
		{
			name: "FloatEQ fails",
			args: []string{"valueTypes", "float", "25"},
			opts: []ArgOpt{
				FloatEQ(24),
			},
			wantErr: "failed to process args: failed to convert value: validation failed: [FloatEQ] value isn't equal to 24",
		},
		// FloatNE
		{
			name: "FloatNE works",
			args: []string{"valueTypes", "float", "24"},
			opts: []ArgOpt{
				FloatNE(25),
			},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType:  FloatType,
					floatVal: 24,
				},
			},
		},
		{
			name: "FloatNE fails",
			args: []string{"valueTypes", "float", "25"},
			opts: []ArgOpt{
				FloatNE(25),
			},
			wantErr: "failed to process args: failed to convert value: validation failed: [FloatNE] value isn't not equal to 25",
		},
		// FloatLT
		{
			name: "FloatLT works",
			args: []string{"valueTypes", "float", "24"},
			opts: []ArgOpt{
				FloatLT(25),
			},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType:  FloatType,
					floatVal: 24,
				},
			},
		},
		{
			name: "FloatLT fails when equal",
			args: []string{"valueTypes", "float", "25"},
			opts: []ArgOpt{
				FloatLT(25),
			},
			wantErr: "failed to process args: failed to convert value: validation failed: [FloatLT] value isn't less than 25",
		},
		{
			name: "FloatLT fails when not less",
			args: []string{"valueTypes", "float", "26"},
			opts: []ArgOpt{
				FloatLT(25),
			},
			wantErr: "failed to process args: failed to convert value: validation failed: [FloatLT] value isn't less than 25",
		},
		// FloatLTE
		{
			name: "FloatLTE works",
			args: []string{"valueTypes", "float", "24"},
			opts: []ArgOpt{
				FloatLTE(25),
			},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType:  FloatType,
					floatVal: 24,
				},
			},
		},
		{
			name: "FloatLTE works when equal",
			args: []string{"valueTypes", "float", "25"},
			opts: []ArgOpt{
				FloatLTE(25),
			},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType:  FloatType,
					floatVal: 25,
				},
			},
		},
		{
			name: "FloatLT fails when not less",
			args: []string{"valueTypes", "float", "26"},
			opts: []ArgOpt{
				FloatLTE(25),
			},
			wantErr: "failed to process args: failed to convert value: validation failed: [FloatLTE] value isn't less than or equal to 25",
		},
		// FloatGT
		{
			name: "FloatGT fails when not greater",
			args: []string{"valueTypes", "float", "24"},
			opts: []ArgOpt{
				FloatGT(25),
			},
			wantErr: "failed to process args: failed to convert value: validation failed: [FloatGT] value isn't greater than 25",
		},
		{
			name: "FloatGT fails when equal",
			args: []string{"valueTypes", "float", "25"},
			opts: []ArgOpt{
				FloatGT(25),
			},
			wantErr: "failed to process args: failed to convert value: validation failed: [FloatGT] value isn't greater than 25",
		},
		{
			name: "FloatGT works",
			args: []string{"valueTypes", "float", "26"},
			opts: []ArgOpt{
				FloatGT(25),
			},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType:  FloatType,
					floatVal: 26,
				},
			},
		},
		// FloatGTE
		{
			name: "FloatGTE fails when not greater",
			args: []string{"valueTypes", "float", "24"},
			opts: []ArgOpt{
				FloatGTE(25),
			},
			wantErr: "failed to process args: failed to convert value: validation failed: [FloatGTE] value isn't greater than or equal to 25",
		},
		{
			name: "FloatGTE works when equal",
			args: []string{"valueTypes", "float", "25"},
			opts: []ArgOpt{
				FloatGTE(25),
			},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType:  FloatType,
					floatVal: 25,
				},
			},
		},
		{
			name: "FloatGTE works",
			args: []string{"valueTypes", "float", "26"},
			opts: []ArgOpt{
				FloatGTE(25),
			},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType:  FloatType,
					floatVal: 26,
				},
			},
		},
		// FloatPositive
		{
			name: "FloatPositive fails when negative",
			args: []string{"valueTypes", "float", "-1"},
			opts: []ArgOpt{
				FloatPositive(),
			},
			wantErr: "failed to process args: failed to convert value: validation failed: [FloatPositive] value isn't positive",
		},
		{
			name: "FloatPositive fails when zero",
			args: []string{"valueTypes", "float", "0"},
			opts: []ArgOpt{
				FloatPositive(),
			},
			wantErr: "failed to process args: failed to convert value: validation failed: [FloatPositive] value isn't positive",
		},
		{
			name: "FloatPositive works when positive",
			args: []string{"valueTypes", "float", "1"},
			opts: []ArgOpt{
				FloatPositive(),
			},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType:  FloatType,
					floatVal: 1,
				},
			},
		},
		// FloatNegative
		{
			name: "FloatNegative works when negative",
			args: []string{"valueTypes", "float", "-1"},
			opts: []ArgOpt{
				FloatNegative(),
			},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType:  FloatType,
					floatVal: -1,
				},
			},
		},
		{
			name: "FloatNegative fails when zero",
			args: []string{"valueTypes", "float", "0"},
			opts: []ArgOpt{
				FloatNegative(),
			},
			wantErr: "failed to process args: failed to convert value: validation failed: [FloatNegative] value isn't negative",
		},
		{
			name: "FloatNegative fails when positive",
			args: []string{"valueTypes", "float", "1"},
			opts: []ArgOpt{
				FloatNegative(),
			},
			wantErr: "failed to process args: failed to convert value: validation failed: [FloatNegative] value isn't negative",
		},
		// FloatNonNegative
		{
			name: "FloatNonNegative fails when negative",
			args: []string{"valueTypes", "float", "-1"},
			opts: []ArgOpt{
				FloatNonNegative(),
			},
			wantErr: "failed to process args: failed to convert value: validation failed: [FloatNonNegative] value isn't non-negative",
		},
		{
			name: "FloatNonNegative works when zero",
			args: []string{"valueTypes", "float", "0"},
			opts: []ArgOpt{
				FloatNonNegative(),
			},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType:  FloatType,
					floatVal: 0,
				},
			},
		},
		{
			name: "FloatNonNegative works when positive",
			args: []string{"valueTypes", "float", "1"},
			opts: []ArgOpt{
				FloatNonNegative(),
			},
			wantExecuteArgs: map[string]*Value{
				"req": &Value{
					valType:  FloatType,
					floatVal: 1,
				},
			},
		},
		/* Useful comment for commenting out tests */
	} {
		t.Run(test.name, func(t *testing.T) {
			var gotExecuteArgs map[string]*Value
			var gotExecuteFlags map[string]*Value

			ex := test.ex
			if ex == nil {
				ex = func(args map[string]*Value, flags map[string]*Value) (*ExecutorResponse, error) {
					// Check length so we can consider empty to be the same as nil.
					// That makes for cleaner test cases.
					if len(args) > 0 {
						gotExecuteArgs = args
					}
					if len(flags) > 0 {
						gotExecuteFlags = flags
					}
					return test.exResp, test.exErr
				}
			}

			cmd := branchCommand(ex, &Completor{}, test.opts...)

			got, err := Execute(cmd, test.args)

			if err != nil && test.wantErr == "" {
				t.Fatalf("command.Execute(%v) returned an error (%v); want nil", test.args, err)
			} else if err == nil && test.wantErr != "" {
				t.Fatalf("command.Execute(%v) returned nil; want error with message %q", test.args, test.wantErr)
			} else if err != nil && !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("command.Execute(%v) returned error (%v); want error with message %q", test.args, err, test.wantErr)
			}

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("command.Execute(%v) returned diff (-want, +got):\n%s", test.args, diff)
			}

			opt := cmp.AllowUnexported(Value{})

			if diff := cmp.Diff(test.wantExecuteArgs, gotExecuteArgs, opt); diff != "" {
				t.Errorf("command.Execute(%v) produced execute args diff (-want, +got):\n%s", test.args, diff)
			}

			if diff := cmp.Diff(test.wantExecuteFlags, gotExecuteFlags, opt); diff != "" {
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
			wantValue: &Value{stringList: []string{""}},
			wantCompleteArgs: map[string]*Value{
				"pos_1": &Value{
					stringList: []string{""},
				},
			},
		},
		{
			name:      "partially complete second arg gets autocompleted",
			args:      []string{"basic", "t"},
			fetchResp: []string{"build", "test", "try", "trying"},
			want:      []string{"test", "try", "trying"},
			wantValue: &Value{stringList: []string{"t"}},
			wantCompleteArgs: map[string]*Value{
				"pos_1": &Value{
					stringList: []string{"t"},
				},
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
			wantValue: &Value{stringList: []string{"o"}},
			wantCompleteArgs: map[string]*Value{
				"pos_1": &Value{
					stringList: []string{"build"},
				},
				"place 2": &Value{
					stringList: []string{"o"},
				},
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
			wantValue: &Value{valType: StringType},
			wantCompleteArgs: map[string]*Value{
				"aight": &Value{valType: StringType},
			},
		},
		// Test lists
		{
			name:      "completes all list suggestions for first one",
			args:      []string{"advanced", "liszt", ""},
			fetchResp: []string{"harp", "piano", "picolo"},
			want:      []string{"harp", "piano", "picolo"},
			wantValue: &Value{stringList: []string{""}},
			wantCompleteArgs: map[string]*Value{
				"list-arg": &Value{
					stringList: []string{""},
				},
			},
		},
		{
			name:      "completes all list suggestions for later one",
			args:      []string{"advanced", "liszt", "un", "deux", "trois", "quatre", "p"},
			fetchResp: []string{"harp", "piano", "picolo"},
			want:      []string{"piano", "picolo"},
			wantValue: &Value{stringList: []string{"un", "deux", "trois", "quatre", "p"}},
			wantCompleteArgs: map[string]*Value{
				"list-arg": &Value{
					stringList: []string{"un", "deux", "trois", "quatre", "p"},
				},
			},
		},
		// Test extra optional arguments
		{
			name:      "optional argument recommends for minimum",
			args:      []string{"sometimes", ""},
			distinct:  true,
			fetchResp: []string{"occ", "occasional", "temp", "temporary", "tmp"},
			want:      []string{"occ", "occasional", "temp", "temporary", "tmp"},
			wantValue: &Value{stringList: []string{""}},
			wantCompleteArgs: map[string]*Value{
				"opt group": &Value{
					stringList: []string{""},
				},
			},
		},
		{
			name:      "optional argument recommends for middle",
			args:      []string{"sometimes", "tmp", "occ", "t"},
			distinct:  true,
			fetchResp: []string{"occ", "occasional", "temp", "temporary", "tmp"},
			want:      []string{"temp", "temporary"},
			wantValue: &Value{stringList: []string{"tmp", "occ", "t"}},
			wantCompleteArgs: map[string]*Value{
				"opt group": &Value{
					stringList: []string{"tmp", "occ", "t"},
				},
			},
		},
		{
			name:      "ignore already listed items",
			args:      []string{"sometimes", "tmp", "occ", "temp", ""},
			distinct:  true,
			fetchResp: []string{"occ", "occasional", "temp", "temporary", "tmp"},
			want:      []string{"occasional", "temporary"},
			wantValue: &Value{stringList: []string{"tmp", "occ", "temp", ""}},
			wantCompleteArgs: map[string]*Value{
				"opt group": &Value{
					stringList: []string{"tmp", "occ", "temp", ""},
				},
			},
		},
		{
			name:      "optional argument recommends for end",
			args:      []string{"sometimes", "tmp", "occ", "temporary", "o"},
			distinct:  true,
			fetchResp: []string{"occ", "occasional", "temp", "temporary", "tmp"},
			want:      []string{"occasional"},
			wantValue: &Value{stringList: []string{"tmp", "occ", "temporary", "o"}},
			wantCompleteArgs: map[string]*Value{
				"opt group": &Value{
					stringList: []string{"tmp", "occ", "temporary", "o"},
				},
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
			wantValue: &Value{stringList: []string{""}},
			wantCompleteArgs: map[string]*Value{
				"alpha": &Value{
					stringList: []string{""},
				},
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
			wantValue: &Value{stringList: []string{"Fo"}},
			wantCompleteArgs: map[string]*Value{
				"alpha": &Value{
					stringList: []string{"Fo"},
				},
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
			wantValue: &Value{stringList: []string{"F"}},
			wantCompleteArgs: map[string]*Value{
				"alpha": &Value{
					stringList: []string{"F"},
				},
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
			wantValue: &Value{stringList: []string{"Greg's One", ""}},
			wantCompleteArgs: map[string]*Value{
				"whose": &Value{
					stringList: []string{"Greg's One", ""},
				},
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
			wantValue: &Value{stringList: []string{`Greg"s Other"s`, ""}},
			wantCompleteArgs: map[string]*Value{
				"whose": &Value{
					stringList: []string{`Greg"s Other"s`, ""},
				},
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
			wantValue: &Value{stringList: []string{""}},
			wantCompleteArgs: map[string]*Value{
				"alpha": &Value{
					stringList: []string{""},
				},
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
			wantValue: &Value{stringList: []string{"First Choice", ""}},
			wantCompleteArgs: map[string]*Value{
				"alpha": &Value{
					stringList: []string{"First Choice", ""},
				},
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
			wantValue: &Value{stringList: []string{"First Choice", "F"}},
			wantCompleteArgs: map[string]*Value{
				"alpha": &Value{
					stringList: []string{"First Choice", "F"},
				},
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
			wantValue: &Value{stringList: []string{""}},
			wantCompleteArgs: map[string]*Value{
				"alpha": &Value{
					stringList: []string{""},
				},
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
			wantValue: &Value{stringList: []string{"F"}},
			wantCompleteArgs: map[string]*Value{
				"alpha": &Value{
					stringList: []string{"F"},
				},
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
			wantValue: &Value{stringList: []string{"Greg's T"}},
			wantCompleteArgs: map[string]*Value{
				"whose": &Value{
					stringList: []string{"Greg's T"},
				},
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
			wantValue: &Value{stringList: []string{""}},
			wantCompleteArgs: map[string]*Value{
				"alpha": &Value{
					stringList: []string{""},
				},
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
			wantValue: &Value{stringList: []string{"F"}},
			wantCompleteArgs: map[string]*Value{
				"alpha": &Value{
					stringList: []string{"F"},
				},
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
			wantValue: &Value{stringList: []string{`Greg"s T`}},
			wantCompleteArgs: map[string]*Value{
				"whose": &Value{
					stringList: []string{`Greg"s T`},
				},
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
			wantValue: &Value{stringList: []string{"Attempt One "}},
			wantCompleteArgs: map[string]*Value{
				"alphas": &Value{
					stringList: []string{"Attempt One "},
				},
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
			wantValue: &Value{stringList: []string{"Three"}},
			wantCompleteArgs: map[string]*Value{
				"alphas": &Value{
					stringList: []string{"Three"},
				},
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
			wantValue: &Value{stringList: []string{"First O"}},
			wantCompleteArgs: map[string]*Value{
				"alpha": &Value{
					stringList: []string{"First O"},
				},
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
			wantValue: &Value{stringList: []string{"e"}},
			wantCompleteArgs: map[string]*Value{
				"syllable": &Value{
					stringList: []string{"e"},
				},
			},
			wantCompleteFlags: map[string]*Value{
				"american": &Value{
					stringList: []string{},
				},
			},
		},
		{
			name:      "regular completion when short boolean flag is earlier",
			args:      []string{"intermediate", "-a", "e"},
			fetchResp: []string{"int", "erm", "edi", "ate"},
			want:      []string{"edi", "erm"},
			wantValue: &Value{stringList: []string{"e"}},
			wantCompleteArgs: map[string]*Value{
				"syllable": &Value{
					stringList: []string{"e"},
				},
			},
			wantCompleteFlags: map[string]*Value{
				"american": &Value{
					stringList: []string{},
				},
			},
		},
		{
			name:      "regular completion when flag with argument is earlier",
			args:      []string{"intermediate", "ate", "--state", "maine", "e"},
			fetchResp: []string{"int", "erm", "edi", "ate"},
			want:      []string{"edi", "erm"},
			wantValue: &Value{stringList: []string{"ate", "e"}},
			wantCompleteArgs: map[string]*Value{
				// TODO: should complete args include the last incomplete variable?
				"syllable": &Value{
					stringList: []string{"ate", "e"},
				},
			},
			wantCompleteFlags: map[string]*Value{
				"state": &Value{
					stringList: []string{"maine"},
				},
			},
		},
		{
			name:      "regular completion when short flag with argument is earlier",
			args:      []string{"intermediate", "ate", "-s", "maine", "e"},
			fetchResp: []string{"int", "erm", "edi", "ate"},
			want:      []string{"edi", "erm"},
			wantValue: &Value{stringList: []string{"ate", "e"}},
			wantCompleteArgs: map[string]*Value{
				"syllable": &Value{
					stringList: []string{"ate", "e"},
				},
			},
			wantCompleteFlags: map[string]*Value{
				"state": &Value{
					stringList: []string{"maine"},
				},
			},
		},
		{
			name:      "flag arguments are autocompleted",
			args:      []string{"intermediate", "--state", ""},
			fetchResp: []string{"california", "connecticut", "washington", "washington_dc"},
			want:      []string{"california", "connecticut", "washington", "washington_dc"},
			wantValue: &Value{stringList: []string{""}},
			wantCompleteFlags: map[string]*Value{
				"state": &Value{
					stringList: []string{""},
				},
			},
		},
		{
			name:      "partial flag arguments are autocompleted",
			args:      []string{"intermediate", "--state", ""},
			fetchResp: []string{"california", "connecticut", "washington", "washington_dc"},
			want:      []string{"california", "connecticut", "washington", "washington_dc"},
			wantValue: &Value{stringList: []string{""}},
			wantCompleteFlags: map[string]*Value{
				"state": &Value{
					stringList: []string{""},
				},
			},
		},
		{
			name:      "short flag arguments are autocompleted",
			args:      []string{"intermediate", "-s", "washington"},
			fetchResp: []string{"california", "connecticut", "washington", "washington_dc"},
			want:      []string{"washington", "washington_dc"},
			wantValue: &Value{stringList: []string{"washington"}},
			wantCompleteFlags: map[string]*Value{
				"state": &Value{
					stringList: []string{"washington"},
				},
			},
		},
		{
			name:      "flag completion works when several flags",
			args:      []string{"intermediate", "--another", "a", "int", "erm", "-a", "edi", "--state", ""},
			fetchResp: []string{"california", "connecticut", "washington", "washington_dc"},
			want:      []string{"california", "connecticut", "washington", "washington_dc"},
			wantValue: &Value{stringList: []string{""}},
			wantCompleteFlags: map[string]*Value{
				"american": &Value{
					stringList: []string{},
				},
				"another": &Value{
					stringList: []string{"a"},
				},
				"state": &Value{
					stringList: []string{""},
				},
			},
		},
		{
			name:      "flag completion works when several flags and partial flag arg",
			args:      []string{"intermediate", "--another", "a", "int", "erm", "-a", "edi", "--state", "wash"},
			fetchResp: []string{"california", "connecticut", "washington", "washington_dc"},
			want:      []string{"washington", "washington_dc"},
			wantValue: &Value{stringList: []string{"wash"}},
			wantCompleteFlags: map[string]*Value{
				"american": &Value{
					stringList: []string{},
				},
				"another": &Value{
					stringList: []string{"a"},
				},
				"state": &Value{
					stringList: []string{"wash"},
				},
			},
		},
		{
			name:      "flag completion works when flag has multiple arguments",
			args:      []string{"wave", "--yourFlag", ""},
			fetchResp: []string{"please", "person", "okay"},
			want:      []string{"okay", "person", "please"},
			wantValue: &Value{
				stringList: []string{""},
			},
			wantCompleteFlags: map[string]*Value{
				"yourFlag": &Value{
					stringList: []string{""},
				},
			},
		},
		{
			name:      "flag partial completion works when flag has multiple arguments",
			args:      []string{"wave", "--yourFlag", "p"},
			fetchResp: []string{"please", "person", "okay"},
			want:      []string{"person", "please"},
			wantValue: &Value{
				stringList: []string{"p"},
			},
			wantCompleteFlags: map[string]*Value{
				"yourFlag": &Value{
					stringList: []string{"p"},
				},
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
				"cb-command": &Value{
					stringList: []string{""},
				},
			},
			wantValue: &Value{stringList: []string{""}},
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
				"cb-command": &Value{
					stringList: []string{"f"},
				},
			},
			wantValue: &Value{stringList: []string{"f"}},
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
				"cb-command": &Value{
					stringList: []string{"noMatch", ""},
				},
			},
			wantValue: &Value{stringList: []string{"noMatch", ""}},
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
				"cb-command": &Value{
					stringList: []string{"noMatch", "for"},
				},
			},
			wantValue: &Value{stringList: []string{"noMatch", "for"}},
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
			wantValue: &Value{
				valType: StringType,
			},
			wantCompleteArgs: map[string]*Value{
				"req": &Value{
					valType: StringType,
				},
			},
		},
		{
			name:      "completes optional string argument",
			args:      []string{"valueTypes", "string", "hello", ""},
			fetchResp: []string{"world", "there", "toYou"},
			want:      []string{"there", "toYou", "world"},
			wantValue: &Value{
				valType: StringType,
			},
			wantCompleteArgs: map[string]*Value{
				"req": &Value{
					valType:   StringType,
					stringVal: "hello",
				},
				"opt": &Value{
					valType: StringType,
				},
			},
		},
		// stringList argument type
		{
			name:      "completes stringList argument",
			args:      []string{"valueTypes", "stringList", "hello", ""},
			fetchResp: []string{"there", "world"},
			want:      []string{"there", "world"},
			wantValue: &Value{
				valType:    StringListType,
				stringList: []string{"hello", ""},
			},
			wantCompleteArgs: map[string]*Value{
				"req": &Value{
					valType:    StringListType,
					stringList: []string{"hello", ""},
				},
			},
		},
		{
			name:      "completes optional stringList argument",
			args:      []string{"valueTypes", "stringList", "hello", "to", ""},
			fetchResp: []string{"them", "you"},
			want:      []string{"them", "you"},
			wantValue: &Value{
				valType:    StringListType,
				stringList: []string{"hello", "to", ""},
			},
			wantCompleteArgs: map[string]*Value{
				"req": &Value{
					valType:    StringListType,
					stringList: []string{"hello", "to", ""},
				},
			},
		},
		// int argument type
		{
			name:      "completes int argument",
			args:      []string{"valueTypes", "int", ""},
			fetchResp: []string{"123", "456"},
			want:      []string{"123", "456"},
			wantValue: &Value{
				valType: IntType,
			},
			wantCompleteArgs: map[string]*Value{
				"req": &Value{
					valType: IntType,
				},
			},
		},
		{
			name:      "completes optional int argument",
			args:      []string{"valueTypes", "int", "123", ""},
			fetchResp: []string{"45", "678"},
			want:      []string{"45", "678"},
			wantValue: &Value{
				valType: IntType,
			},
			wantCompleteArgs: map[string]*Value{
				"req": &Value{
					valType: IntType,
					intVal:  123,
				},
				"opt": &Value{
					valType: IntType,
				},
			},
		},
		{
			name:      "completes when previous int was bad format",
			args:      []string{"valueTypes", "int", "123.45", ""},
			fetchResp: []string{"45", "678"},
			want:      []string{"45", "678"},
			wantValue: &Value{
				valType: IntType,
			},
			wantCompleteArgs: map[string]*Value{
				"req": &Value{
					valType: IntType,
				},
				"opt": &Value{
					valType: IntType,
				},
			},
		},
		// intList argument type
		{
			name:      "completes intList argument",
			args:      []string{"valueTypes", "intList", "123", ""},
			fetchResp: []string{"45", "678"},
			want:      []string{"45", "678"},
			wantValue: &Value{
				valType: IntListType,
				intList: []int{123, 0},
			},
			wantCompleteArgs: map[string]*Value{
				"req": &Value{
					valType: IntListType,
					intList: []int{123, 0},
				},
			},
		},
		{
			name:      "completes optional intList argument",
			args:      []string{"valueTypes", "intList", "123", "45", ""},
			fetchResp: []string{"67", "89"},
			want:      []string{"67", "89"},
			wantValue: &Value{
				valType: IntListType,
				intList: []int{123, 45, 0},
			},
			wantCompleteArgs: map[string]*Value{
				"req": &Value{
					valType: IntListType,
					intList: []int{123, 45, 0},
				},
			},
		},
		{
			name:      "completes intList when previous argument is invalid",
			args:      []string{"valueTypes", "intList", "twelve", "45", ""},
			fetchResp: []string{"67", "89"},
			want:      []string{"67", "89"},
			wantValue: &Value{
				valType: IntListType,
				intList: []int{0, 45, 0},
			},
			wantCompleteArgs: map[string]*Value{
				"req": &Value{
					valType: IntListType,
					intList: []int{0, 45, 0},
				},
			},
		},
		// float argument type
		{
			name:      "completes float argument",
			args:      []string{"valueTypes", "float", ""},
			fetchResp: []string{"12.3", "-456"},
			want:      []string{"-456", "12.3"},
			wantValue: &Value{
				valType: FloatType,
			},
			wantCompleteArgs: map[string]*Value{
				"req": &Value{
					valType: FloatType,
				},
			},
		},
		{
			name:      "completes optional float argument",
			args:      []string{"valueTypes", "float", "1.23", ""},
			fetchResp: []string{"-4.5", "678"},
			want:      []string{"-4.5", "678"},
			wantValue: &Value{
				valType: FloatType,
			},
			wantCompleteArgs: map[string]*Value{
				"req": &Value{
					valType:  FloatType,
					floatVal: 1.23,
				},
				"opt": &Value{
					valType: FloatType,
				},
			},
		},
		{
			name:      "completes when previous float was bad format",
			args:      []string{"valueTypes", "float", "eleven", ""},
			fetchResp: []string{"-45", "67.8"},
			want:      []string{"-45", "67.8"},
			wantValue: &Value{
				valType: FloatType,
			},
			wantCompleteArgs: map[string]*Value{
				"req": &Value{
					valType: FloatType,
				},
				"opt": &Value{
					valType: FloatType,
				},
			},
		},
		// floatList argument type
		{
			name:      "completes floatList argument",
			args:      []string{"valueTypes", "floatList", "123.", ""},
			fetchResp: []string{"4.5", "-678."},
			want:      []string{"-678.", "4.5"},
			wantValue: &Value{
				valType:   FloatListType,
				floatList: []float64{123, 0},
			},
			wantCompleteArgs: map[string]*Value{
				"req": &Value{
					valType:   FloatListType,
					floatList: []float64{123, 0},
				},
			},
		},
		{
			name:      "completes optional floatList argument",
			args:      []string{"valueTypes", "floatList", "0.123", "-.45", ""},
			fetchResp: []string{".67", "-.89"},
			want:      []string{"-.89", ".67"},
			wantValue: &Value{
				valType:   FloatListType,
				floatList: []float64{0.123, -0.45, 0},
			},
			wantCompleteArgs: map[string]*Value{
				"req": &Value{
					valType:   FloatListType,
					floatList: []float64{0.123, -0.45, 0},
				},
			},
		},
		{
			// TODO: test distinct (fetchResp return list of values instead?)
			name:      "completes floatList when previous argument is invalid",
			args:      []string{"valueTypes", "floatList", "twelve", "6.7", ""},
			fetchResp: []string{"6.7", "89"},
			distinct:  true,
			want:      []string{"6.7", "89"},
			wantValue: &Value{
				valType:   FloatListType,
				floatList: []float64{0, 6.7, 0},
			},
			wantCompleteArgs: map[string]*Value{
				"req": &Value{
					valType:   FloatListType,
					floatList: []float64{0, 6.7, 0},
				},
			},
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

			opt := cmp.AllowUnexported(Value{})

			if diff := cmp.Diff(test.wantCompleteArgs, fetcher.gotArgs, opt); diff != "" {
				t.Errorf("command.Autocomplete(%v, %d) produced complete args diff (-want +got):\n%s", test.args, test.cursorIdx, diff)
			}

			if diff := cmp.Diff(test.wantCompleteFlags, fetcher.gotFlags, opt); diff != "" {
				t.Errorf("command.Autocomplete(%v, %d) produced complete flags diff (-want +got):\n%s", test.args, test.cursorIdx, diff)
			}

			if diff := cmp.Diff(test.wantValue, fetcher.gotValue, opt); diff != "" {
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
			"a": &Value{
				stringList: []string{"b"},
			},
			"c": &Value{
				stringList: []string{"d", "e"},
			},
		}
		flags := map[string]*Value{
			"f0": &Value{
				stringList: []string{},
			},
			"f1": &Value{
				stringList: []string{"12", "3"},
			},
			"f4": &Value{
				stringList: []string{"4", "56"},
			},
		}

		if resp, err := NoopExecutor(args, flags); resp != nil && err != nil {
			t.Errorf("Expected NoopExecutor to return (nil, nil); got (%v, %v)", resp, err)
		}
	})

	t.Run("boolean flag has no options", func(t *testing.T) {
		bf := NewBooleanFlag("truther", 't', true)
		as := map[string]*Value{
			"a": &Value{
				stringList: []string{"b"},
			},
		}
		fs := map[string]*Value{
			"c": &Value{
				stringList: []string{"d", "e"},
			},
		}
		if opts := bf.Complete(as, fs); opts != nil {
			t.Errorf("Expected boolean flag options to be nil; got %v", opts)
		}
	})

	t.Run("completor with nil fetch options", func(t *testing.T) {
		c := &Completor{}
		v := &Value{stringList: []string{"he", "yo"}}
		as := map[string]*Value{
			"hey": &Value{
				stringList: []string{"oooo"},
			},
		}
		fs := map[string]*Value{
			"hey": &Value{
				stringList: []string{"o", "o"},
			},
		}
		_ = c.Complete(v, as, fs)
	})

	t.Run("invalid value type throws an error", func(t *testing.T) {
		ap := argProcessor{
			ValueType: -123,
			MinN:      1,
		}

		arg := []string{"1"}
		got, err := ap.Value(arg)
		if got != nil {
			t.Errorf("argProcessor.Value(%s) returned %v; want nil", arg, got)
		}

		if err == nil {
			t.Errorf("argProcessor.Value(%s) returned nil err; want error", arg)
		}

		wantErr := "invalid value type: -123"
		if !strings.Contains(err.Error(), wantErr) {
			t.Errorf("argProcessor.Value(%s) returned err (%v); want error with message %q", arg, err, wantErr)
		}
	})
}
