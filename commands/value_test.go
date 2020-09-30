package commands

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func stringP(s string) *string {
	return &s
}

func stringListP(sl []string) *[]string {
	return &sl
}

func intP(i int) *int {
	return &i
}

func intListP(il []int) *[]int {
	return &il
}

func floatP(f float64) *float64 {
	return &f
}

func floatListP(fl []float64) *[]float64 {
	return &fl
}

func TestValueComands(t *testing.T) {
	for _, test := range []struct {
		name           string
		vt             ValueType
		argDef         Arg
		args           []string
		wantString     *string
		wantStringList *[]string
		wantInt        *int
		wantIntList    *[]int
		wantFloat      *float64
		wantFloatList  *[]float64
		wantErr        string
		want           *ExecutorResponse
	}{
		{
			name:       "string is populated",
			vt:         StringType,
			argDef:     StringArg("argName", true, nil),
			args:       []string{"string-val"},
			wantString: stringP("string-val"),
			want:       &ExecutorResponse{},
		},
		{
			name:           "string list is populated",
			vt:             StringListType,
			argDef:         StringListArg("argName", 2, 3, nil),
			args:           []string{"string", "list", "val"},
			wantStringList: &[]string{"string", "list", "val"},
			want:           &ExecutorResponse{},
		},
		{
			name:    "int is populated",
			vt:      IntType,
			argDef:  IntArg("argName", true, nil),
			args:    []string{"123"},
			wantInt: intP(123),
			want:    &ExecutorResponse{},
		},
		{
			name:        "int list is populated",
			vt:          IntListType,
			argDef:      IntListArg("argName", 2, 3, nil),
			args:        []string{"12", "345", "6"},
			wantIntList: &[]int{12, 345, 6},
			want:        &ExecutorResponse{},
		},
		{
			name:      "flaot is populated",
			vt:        FloatType,
			argDef:    FloatArg("argName", true, nil),
			args:      []string{"12.3"},
			wantFloat: floatP(12.3),
			want:      &ExecutorResponse{},
		},
		{
			name:          "float list is populated",
			vt:            FloatListType,
			argDef:        FloatListArg("argName", 2, 3, nil),
			args:          []string{"1.2", "-345", ".6"},
			wantFloatList: &[]float64{1.2, -345, .6},
			want:          &ExecutorResponse{},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			cmd := &TerminusCommand{
				Args: []Arg{
					test.argDef,
				},
				Executor: func(args, flags map[string]*Value) (*ExecutorResponse, error) {
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

					return &ExecutorResponse{}, nil
				},
			}

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
		})
	}
}
