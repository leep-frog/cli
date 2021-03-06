package commands

import (
	"fmt"
	"strings"
)

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
	case *Value_String_:
		return vt == StringType
	case *Value_Int:
		return vt == IntType
	case *Value_Float:
		return vt == FloatType
	case *Value_Bool:
		return vt == BoolType
	case *Value_StringList:
		return vt == StringListType
	case *Value_IntList:
		return vt == IntListType
	case *Value_FloatList:
		return vt == FloatListType
	}
	return false
}

func (v *Value) Str() string {
	switch v.Type.(type) {
	case *Value_String_:
		return v.GetString_()
	case *Value_Int:
		return fmt.Sprintf(intFmt, v.GetInt())
	case *Value_Float:
		return fmt.Sprintf(floatFmt, v.GetFloat())
	case *Value_Bool:
		if v.GetBool() {
			return "true"
		}
		return "false"
	case *Value_StringList:
		return strings.Join(v.GetStringList().GetList(), ", ")
	case *Value_IntList:
		return intSliceToString(v.GetIntList().GetList())
	case *Value_FloatList:
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
	case *Value_StringList:
		return len(v.GetStringList().GetList())
	case *Value_IntList:
		return len(v.GetIntList().GetList())
	case *Value_FloatList:
		return len(v.GetFloatList().GetList())
	case nil:
		// The field is not set.
		return 0
	}

	// The field is set and is a singular.
	return 1
}
