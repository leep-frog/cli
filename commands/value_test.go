package commands

import (
	"encoding/json"
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
		wantInt        int
		wantIntList    []int
		wantFloat      float64
		wantFloatList  []float64
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
			wantIntList: []int{12, 345, 6},
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
			wantFloatList: []float64{1.2, -345, .6},
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
					if diff := cmp.Diff(test.wantString, v.String()); diff != "" {
						t.Errorf("String() produced diff (-want, +got):\n%s", diff)
					}
					if diff := cmp.Diff(test.wantStringList, v.StringList()); diff != "" {
						t.Errorf("StringList() produced diff (-want, +got):\n%s", diff)
					}

					// ints
					if diff := cmp.Diff(test.wantInt, v.Int()); diff != "" {
						t.Errorf("Int() produced diff (-want, +got):\n%s", diff)
					}
					if diff := cmp.Diff(test.wantIntList, v.IntList()); diff != "" {
						t.Errorf("IntList() produced diff (-want, +got):\n%s", diff)
					}

					// floats
					if diff := cmp.Diff(test.wantFloat, v.Float()); diff != "" {
						t.Errorf("Float() produced diff (-want, +got):\n%s", diff)
					}
					if diff := cmp.Diff(test.wantFloatList, v.FloatList()); diff != "" {
						t.Errorf("FloatList() produced diff (-want, +got):\n%s", diff)
					}

					// bool
					if diff := cmp.Diff(test.wantBool, v.Bool()); diff != "" {
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

func TestValueStrAndJson(t *testing.T) {
	for _, test := range []struct {
		name    string
		v       *Value
		wantStr string
	}{
		{
			name:    "string value",
			v:       StringValue("hello there"),
			wantStr: "hello there",
		},
		{
			name:    "int value",
			v:       IntValue(12),
			wantStr: "12",
		},
		{
			name:    "float value with extra decimal points",
			v:       FloatValue(123.4567),
			wantStr: "123.46",
		},
		{
			name:    "float value with no decimal points",
			v:       FloatValue(123),
			wantStr: "123.00",
		},
		{
			name:    "bool true value",
			v:       BoolValue(true),
			wantStr: "true",
		},
		{
			name:    "bool false value",
			v:       BoolValue(false),
			wantStr: "false",
		},
		{
			name:    "string list",
			v:       StringListValue("hello", "there"),
			wantStr: "hello, there",
		},
		{
			name:    "int list",
			v:       IntListValue(12, -34, 5678),
			wantStr: "12, -34, 5678",
		},
		{
			name:    "float list",
			v:       FloatListValue(0.12, -3.4, 567.8910),
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

func TestValueEqualAndJSONMarshaling(t *testing.T) {
	for _, test := range []struct {
		name         string
		this         *Value
		that         *Value
		wantThisJSON string
		wantThatJSON string
		want         bool
	}{
		{
			name:         "nil values are equal",
			want:         true,
			wantThisJSON: "null",
			wantThatJSON: "null",
		},
		{
			name:         "nil vs not nil aren't equal",
			this:         StringValue(""),
			wantThisJSON: `{"Type":2,"String":""}`,
			wantThatJSON: "null",
		},
		{
			name:         "equal empty string values",
			this:         StringValue(""),
			that:         StringValue(""),
			want:         true,
			wantThisJSON: `{"Type":2,"String":""}`,
			wantThatJSON: `{"Type":2,"String":""}`,
		},
		{
			name:         "equal string values",
			this:         StringValue("this"),
			that:         StringValue("this"),
			wantThisJSON: `{"Type":2,"String":"this"}`,
			wantThatJSON: `{"Type":2,"String":"this"}`,
			want:         true,
		},
		{
			name:         "unequal string values",
			this:         StringValue("this"),
			that:         StringValue("that"),
			wantThisJSON: `{"Type":2,"String":"this"}`,
			wantThatJSON: `{"Type":2,"String":"that"}`,
		},
		{
			name:         "empty equal int values",
			this:         IntValue(0),
			that:         IntValue(0),
			want:         true,
			wantThisJSON: `{"Type":3,"Int":0}`,
			wantThatJSON: `{"Type":3,"Int":0}`,
		},
		{
			name:         "equal int values",
			this:         IntValue(1),
			that:         IntValue(1),
			want:         true,
			wantThisJSON: `{"Type":3,"Int":1}`,
			wantThatJSON: `{"Type":3,"Int":1}`,
		},
		{
			name:         "unequal int values",
			this:         IntValue(0),
			that:         IntValue(1),
			wantThisJSON: `{"Type":3,"Int":0}`,
			wantThatJSON: `{"Type":3,"Int":1}`,
		},
		{
			name:         "empty equal float values",
			this:         FloatValue(0),
			that:         FloatValue(0),
			want:         true,
			wantThisJSON: `{"Type":5,"Float":0}`,
			wantThatJSON: `{"Type":5,"Float":0}`,
		},
		{
			name:         "equal float values",
			this:         FloatValue(2.4),
			that:         FloatValue(2.4),
			want:         true,
			wantThisJSON: `{"Type":5,"Float":2.4}`,
			wantThatJSON: `{"Type":5,"Float":2.4}`,
		},
		{
			name:         "unequal float values",
			this:         FloatValue(1.1),
			that:         FloatValue(2.2),
			wantThisJSON: `{"Type":5,"Float":1.1}`,
			wantThatJSON: `{"Type":5,"Float":2.2}`,
		},
		{
			name:         "equal bool values",
			this:         BoolValue(true),
			that:         BoolValue(true),
			want:         true,
			wantThisJSON: `{"Type":7,"Bool":true}`,
			wantThatJSON: `{"Type":7,"Bool":true}`,
		},
		{
			name:         "unequal bool values",
			this:         BoolValue(true),
			that:         BoolValue(false),
			wantThisJSON: `{"Type":7,"Bool":true}`,
			wantThatJSON: `{"Type":7,"Bool":false}`,
		},
		{
			name:         "empty string list",
			this:         StringListValue(),
			that:         StringListValue(),
			want:         true,
			wantThisJSON: `{"Type":1,"StringList":null}`,
			wantThatJSON: `{"Type":1,"StringList":null}`,
		},
		{
			name:         "unequal empty string list",
			this:         StringListValue("a"),
			that:         StringListValue(),
			wantThisJSON: `{"Type":1,"StringList":["a"]}`,
			wantThatJSON: `{"Type":1,"StringList":null}`,
		},
		{
			name:         "populated string list",
			this:         StringListValue("a", "bc", "d"),
			that:         StringListValue("a", "bc", "d"),
			want:         true,
			wantThisJSON: `{"Type":1,"StringList":["a","bc","d"]}`,
			wantThatJSON: `{"Type":1,"StringList":["a","bc","d"]}`,
		},
		{
			name:         "different string list",
			this:         StringListValue("a", "bc", "def"),
			that:         StringListValue("a", "bc", "d"),
			wantThisJSON: `{"Type":1,"StringList":["a","bc","def"]}`,
			wantThatJSON: `{"Type":1,"StringList":["a","bc","d"]}`,
		},
		{
			name:         "unequal populated string list",
			this:         StringListValue("a", "bc", "d"),
			that:         StringListValue("a", "bc"),
			wantThisJSON: `{"Type":1,"StringList":["a","bc","d"]}`,
			wantThatJSON: `{"Type":1,"StringList":["a","bc"]}`,
		},
		{
			name:         "empty int list",
			this:         IntListValue(),
			that:         IntListValue(),
			want:         true,
			wantThisJSON: `{"Type":4,"IntList":null}`,
			wantThatJSON: `{"Type":4,"IntList":null}`,
		},
		{
			name:         "unequal empty int list",
			this:         IntListValue(0),
			that:         IntListValue(),
			wantThisJSON: `{"Type":4,"IntList":[0]}`,
			wantThatJSON: `{"Type":4,"IntList":null}`,
		},
		{
			name:         "populated int list",
			this:         IntListValue(1, -23, 456),
			that:         IntListValue(1, -23, 456),
			want:         true,
			wantThisJSON: `{"Type":4,"IntList":[1,-23,456]}`,
			wantThatJSON: `{"Type":4,"IntList":[1,-23,456]}`,
		},
		{
			name:         "different int list",
			this:         IntListValue(1, -23, 789),
			that:         IntListValue(1, -23, 456),
			wantThisJSON: `{"Type":4,"IntList":[1,-23,789]}`,
			wantThatJSON: `{"Type":4,"IntList":[1,-23,456]}`,
		},
		{
			name:         "unequal populated int list",
			this:         IntListValue(1, -23, 456),
			that:         IntListValue(1, -23),
			wantThisJSON: `{"Type":4,"IntList":[1,-23,456]}`,
			wantThatJSON: `{"Type":4,"IntList":[1,-23]}`,
		},
		{
			name:         "empty float list",
			this:         FloatListValue(),
			that:         FloatListValue(),
			want:         true,
			wantThisJSON: `{"Type":6,"FloatList":null}`,
			wantThatJSON: `{"Type":6,"FloatList":null}`,
		},
		{
			name:         "unequal empty float list",
			this:         FloatListValue(0),
			that:         FloatListValue(),
			wantThisJSON: `{"Type":6,"FloatList":[0]}`,
			wantThatJSON: `{"Type":6,"FloatList":null}`,
		},
		{
			name:         "populated float list",
			this:         FloatListValue(1, -2.3, 0.456),
			that:         FloatListValue(1, -2.3, 0.456),
			want:         true,
			wantThisJSON: `{"Type":6,"FloatList":[1,-2.3,0.456]}`,
			wantThatJSON: `{"Type":6,"FloatList":[1,-2.3,0.456]}`,
		},
		{
			name:         "different float list",
			this:         FloatListValue(1, -2.3, 45.6),
			that:         FloatListValue(1, -2.3, 0.456),
			wantThisJSON: `{"Type":6,"FloatList":[1,-2.3,45.6]}`,
			wantThatJSON: `{"Type":6,"FloatList":[1,-2.3,0.456]}`,
		},
		{
			name:         "unequal populated float list",
			this:         FloatListValue(1, -2.3, 0.456),
			that:         FloatListValue(-2.3, 0.456),
			wantThisJSON: `{"Type":6,"FloatList":[1,-2.3,0.456]}`,
			wantThatJSON: `{"Type":6,"FloatList":[-2.3,0.456]}`,
		},
		/* Usefor for commenting out tests. */
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := test.this.Equal(test.that); got != test.want {
				t.Errorf("Value(%v).Equal(Value(%v)) returned %v; want %v", test.this, test.that, got, test.want)
			}

			if got := test.that.Equal(test.this); got != test.want {
				t.Errorf("Value(%v).Equal(Value(%v)) returned %v; want %v", test.that, test.this, got, test.want)
			}

			gotThisJSON, err := json.Marshal(test.this)
			if err != nil {
				t.Fatalf("json.Marshal(%v) [this] returned error: %v", test.this, err)
			}
			if diff := cmp.Diff(test.wantThisJSON, string(gotThisJSON)); diff != "" {
				t.Errorf("json.Marshal(%v) [this] produced diff (-want, +got):\n%s", test.this, diff)
			}

			gotThatJSON, err := json.Marshal(test.that)
			if err != nil {
				t.Fatalf("json.Marshal(%v) [that] returned error: %v", test.that, err)
			}
			if diff := cmp.Diff(test.wantThatJSON, string(gotThatJSON)); diff != "" {
				t.Errorf("json.Marshal(%v) [that] produced diff (-want, +got):\n%s", test.that, diff)
			}

			// Unmarshal and verify still equal.
			unmarshalledThis := &Value{}
			if err := json.Unmarshal(gotThisJSON, unmarshalledThis); err != nil {
				t.Fatalf("json.Unmarshal(%v) [this] returned an error: %v", gotThisJSON, err)
			}
			wantThis := test.this
			if test.this == nil {
				wantThis = &Value{}
			}
			if diff := cmp.Diff(wantThis, unmarshalledThis); diff != "" {
				t.Errorf("json marshal + unmarshal [this] produced diff (-want, +got):\n%s", diff)
			}

			unmarshalledThat := &Value{}
			if err := json.Unmarshal(gotThatJSON, unmarshalledThat); err != nil {
				t.Fatalf("json.Unmarshal(%v) [that] returned an error: %v", gotThatJSON, err)
			}
			wantThat := test.that
			if test.that == nil {
				wantThat = &Value{}
			}
			if diff := cmp.Diff(wantThat, unmarshalledThat); diff != "" {
				t.Errorf("json marshal + unmarshal [that] produced diff (-want, +got):\n%s", diff)
			}
		})
	}
}
