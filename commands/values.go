package commands

import (
	"fmt"
	"strings"

	vpb "github.com/leep-frog/commands/commands/value"
)

func StringListValue(s ...string) *Value {
	return &Value{&vpb.Value{
		Type: &vpb.Value_StringList{
			StringList: &vpb.StringList{
				List: s,
			},
		},
		Set: true,
	}}
}

func IntListValue(l ...int32) *Value {
	return &Value{&vpb.Value{
		Type: &vpb.Value_IntList{
			IntList: &vpb.IntList{
				List: l,
			},
		},
		Set: true,
	}}
}

func FloatListValue(l ...float32) *Value {
	return &Value{&vpb.Value{
		Type: &vpb.Value_FloatList{
			FloatList: &vpb.FloatList{
				List: l,
			},
		},
		Set: true,
	}}
}

func BoolValue(b bool) *Value {
	return &Value{&vpb.Value{
		Type: &vpb.Value_Bool{
			Bool: b,
		},
		Set: true,
	}}
}

func StringValue(s string) *Value {
	return &Value{&vpb.Value{
		Type: &vpb.Value_String_{
			String_: s,
		},
		Set: true,
	}}
}

func IntValue(i int32) *Value {
	return &Value{&vpb.Value{
		Type: &vpb.Value_Int{
			Int: i,
		},
		Set: true,
	}}
}

func FloatValue(f float32) *Value {
	return &Value{&vpb.Value{
		Type: &vpb.Value_Float{
			Float: f,
		},
		Set: true,
	}}
}

type Value struct {
	*vpb.Value
}

func (v *Value) Provided() bool {
	return v != nil && v.GetSet()
}

type ValueType int

const (
	// Default is string list
	StringListType ValueType = iota
	StringType
	IntType
	IntListType
	FloatType
	FloatListType
	BoolType

	floatFmt = "%.2f"
	intFmt   = "%d"
)

var (
	boolStringMap = map[string]bool{
		"true":  true,
		"t":     true,
		"false": false,
		"f":     false,
	}
)

func (v *Value) IsType(vt ValueType) bool {
	switch v.Type.(type) {
	case *vpb.Value_String_:
		return vt == StringType
	case *vpb.Value_Int:
		return vt == IntType
	case *vpb.Value_Float:
		return vt == FloatType
	case *vpb.Value_Bool:
		return vt == BoolType
	case *vpb.Value_StringList:
		return vt == StringListType
	case *vpb.Value_IntList:
		return vt == IntListType
	case *vpb.Value_FloatList:
		return vt == FloatListType
	}
	return false
}

func (v *Value) Str() string {
	switch v.Type.(type) {
	case *vpb.Value_String_:
		return v.GetString_()
	case *vpb.Value_Int:
		return fmt.Sprintf(intFmt, v.GetInt())
	case *vpb.Value_Float:
		return fmt.Sprintf(floatFmt, v.GetFloat())
	case *vpb.Value_Bool:
		if v.GetBool() {
			return "true"
		}
		return "false"
	case *vpb.Value_StringList:
		return strings.Join(v.GetStringList().GetList(), ", ")
	case *vpb.Value_IntList:
		return intSliceToString(v.GetIntList().GetList())
	case *vpb.Value_FloatList:
		return floatSliceToString(v.GetFloatList().GetList())
	}
	// Unreachable
	return "UNKNOWN"
}

func intSliceToString(is []int32) string {
	ss := make([]string, 0, len(is))
	for _, i := range is {
		ss = append(ss, fmt.Sprintf("%d", i))
	}
	return strings.Join(ss, ", ")
}

func floatSliceToString(fs []float32) string {
	ss := make([]string, 0, len(fs))
	for _, f := range fs {
		ss = append(ss, fmt.Sprintf(floatFmt, f))
	}
	return strings.Join(ss, ", ")
}

func (v *Value) Length() int {
	switch v.Type.(type) {
	case *vpb.Value_StringList:
		return len(v.GetStringList().GetList())
	case *vpb.Value_IntList:
		return len(v.GetIntList().GetList())
	case *vpb.Value_FloatList:
		return len(v.GetFloatList().GetList())
	case nil:
		// The field is not set.
		return 0
	}

	// The field is set and is a singular.
	return 1
}
