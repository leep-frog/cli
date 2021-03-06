package commands

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestValueCommands(t *testing.T) {
	for _, test := range []struct {
		name           string
		argDef         Arg
		args           []string
		wantString     string
		wantStringList []string
		wantInt        int32
		wantIntList    []int32
		wantFloat      float32
		wantFloatList  []float32
		wantBool       bool
		wantOK         bool
		want           *ExecutorResponse
		wantStdout     []string
		wantStderr     []string
	}{
		{
			name:       "string is populated",
			argDef:     StringArg("argName", true, nil),
			args:       []string{"string-val"},
			wantString: "string-val",
			want:       &ExecutorResponse{},
			wantOK:     true,
		},
		{
			name:           "string list is populated",
			argDef:         StringListArg("argName", 2, 3, nil),
			args:           []string{"string", "list", "val"},
			wantStringList: []string{"string", "list", "val"},
			want:           &ExecutorResponse{},
			wantOK:         true,
		},
		{
			name:    "int is populated",
			argDef:  IntArg("argName", true, nil),
			args:    []string{"123"},
			wantInt: 123,
			want:    &ExecutorResponse{},
			wantOK:  true,
		},
		{
			name:        "int list is populated",
			argDef:      IntListArg("argName", 2, 3, nil),
			args:        []string{"12", "345", "6"},
			wantIntList: []int32{12, 345, 6},
			want:        &ExecutorResponse{},
			wantOK:      true,
		},
		{
			name:      "flaot is populated",
			argDef:    FloatArg("argName", true, nil),
			args:      []string{"12.3"},
			wantFloat: 12.3,
			want:      &ExecutorResponse{},
			wantOK:    true,
		},
		{
			name:          "float list is populated",
			argDef:        FloatListArg("argName", 2, 3, nil),
			args:          []string{"1.2", "-345", ".6"},
			wantFloatList: []float32{1.2, -345, .6},
			want:          &ExecutorResponse{},
			wantOK:        true,
		},
		{
			name:     "bool is populated",
			argDef:   BoolArg("argName", true),
			args:     []string{"true"},
			wantBool: true,
			want:     &ExecutorResponse{},
			wantOK:   true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			cmd := &TerminusCommand{
				Args: []Arg{
					test.argDef,
				},
				Executor: func(cos CommandOS, args, flags map[string]*Value, _ *OptionInfo) (*ExecutorResponse, bool) {
					v := args[test.argDef.Name()]

					// strings
					if diff := cmp.Diff(test.wantString, v.GetString_()); diff != "" {
						t.Errorf("String() produced diff (-want, +got):\n%s", diff)
					}
					if diff := cmp.Diff(test.wantStringList, v.GetStringList().GetList()); diff != "" {
						t.Errorf("StringList() produced diff (-want, +got):\n%s", diff)
					}

					// ints
					if diff := cmp.Diff(test.wantInt, v.GetInt()); diff != "" {
						t.Errorf("Int() produced diff (-want, +got):\n%s", diff)
					}
					if diff := cmp.Diff(test.wantIntList, v.GetIntList().GetList()); diff != "" {
						t.Errorf("IntList() produced diff (-want, +got):\n%s", diff)
					}

					// floats
					if diff := cmp.Diff(test.wantFloat, v.GetFloat()); diff != "" {
						t.Errorf("Float() produced diff (-want, +got):\n%s", diff)
					}
					if diff := cmp.Diff(test.wantFloatList, v.GetFloatList().GetList()); diff != "" {
						t.Errorf("FloatList() produced diff (-want, +got):\n%s", diff)
					}

					// bool
					if diff := cmp.Diff(test.wantBool, v.GetBool()); diff != "" {
						t.Errorf("Bool() produced diff (-want, +got):\n%s", diff)
					}

					return &ExecutorResponse{}, true
				},
			}

			tcos := &TestCommandOS{}
			got, ok := Execute(tcos, cmd, test.args, nil)

			if ok != test.wantOK {
				t.Fatalf("commands.Execute(%v) returned %v for ok; want %v", test.args, ok, test.wantOK)
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
		})
	}
}

func TestStr(t *testing.T) {
	for _, test := range []struct {
		name    string
		v       *Value
		wantStr string
	}{
		{
			name:    "string value",
			v:       stringVal("hello there"),
			wantStr: "hello there",
		},
		{
			name:    "int value",
			v:       intVal(12),
			wantStr: "12",
		},
		{
			name:    "float value with extra decimal points",
			v:       floatVal(123.4567),
			wantStr: "123.46",
		},
		{
			name:    "float value with no decimal points",
			v:       floatVal(123),
			wantStr: "123.00",
		},
		{
			name:    "bool true value",
			v:       boolVal(true),
			wantStr: "true",
		},
		{
			name:    "bool false value",
			v:       boolVal(false),
			wantStr: "false",
		},
		{
			name:    "string list",
			v:       stringList("hello", "there"),
			wantStr: "hello, there",
		},
		{
			name:    "int list",
			v:       intList(12, -34, 5678),
			wantStr: "12, -34, 5678",
		},
		{
			name:    "float list",
			v:       floatList(0.12, -3.4, 567.8910),
			wantStr: "0.12, -3.40, 567.89",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			if diff := cmp.Diff(test.wantStr, test.v.Str()); diff != "" {
				t.Errorf("Value.Str() returned incorrect string (-want, +got):\n%s", diff)
			}

		})
	}
}
