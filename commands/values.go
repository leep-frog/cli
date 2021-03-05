package commands

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
