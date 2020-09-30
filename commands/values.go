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
)

// Value is the populated value.
// TODO: change this to use interfaces for each type?
// Might be difficult because return type has to be something
type Value struct {
	valType ValueType

	// One enumeration for each type.
	stringVal  string
	intVal     int
	stringList []string
	intList    []int
	floatVal   float64
	floatList  []float64
}

func (v *Value) Length() int {
	switch v.valType {
	case StringListType:
		return len(v.stringList)
	case IntListType:
		return len(v.intList)
	case FloatListType:
		return len(v.floatList)
	}
	return 1
}

func (v *Value) String() *string {
	if v == nil || v.valType != StringType {
		return nil
	}
	return &v.stringVal
}

func (v *Value) StringList() *[]string {
	if v == nil || v.valType != StringListType {
		return nil
	}
	return &v.stringList
}

func (v *Value) Int() *int {
	if v == nil || v.valType != IntType {
		return nil
	}
	return &v.intVal
}

func (v *Value) IntList() *[]int {
	if v == nil || v.valType != IntListType {
		return nil
	}
	return &v.intList
}

func (v *Value) Float() *float64 {
	if v == nil || v.valType != FloatType {
		return nil
	}
	return &v.floatVal
}

func (v *Value) FloatList() *[]float64 {
	if v == nil || v.valType != FloatListType {
		return nil
	}
	return &v.floatList
}
