package commands

import (
	"fmt"
	"strings"
)

type option struct {
	vt       ValueType
	validate func(*Value) error
}

func (o *option) ValueType() ValueType {
	return o.vt
}

func (o *option) Validate(v *Value) error {
	return o.validate(v)
}

// String options
func StringOption(f func(string) bool, err error) ArgOpt {
	validator := func(v *Value) error {
		if !f(v.GetString_()) {
			return err
		}
		return nil
	}
	return &option{
		vt:       StringType,
		validate: validator,
	}
}

func Contains(s string) ArgOpt {
	return StringOption(
		func(vs string) bool { return strings.Contains(vs, s) },
		fmt.Errorf("[Contains] value doesn't contain substring %q", s),
	)
}

func MinLength(length int) ArgOpt {
	return StringOption(
		func(vs string) bool { return len(vs) >= length },
		fmt.Errorf("[MinLength] value must be at least %d characters", length),
	)
}

// Int options
func IntOption(f func(int32) bool, err error) ArgOpt {
	validator := func(v *Value) error {
		if !f(v.GetInt()) {
			return err
		}
		return nil
	}
	return &option{
		vt:       IntType,
		validate: validator,
	}
}

func IntEQ(i int32) ArgOpt {
	return IntOption(
		func(vi int32) bool { return vi == i },
		fmt.Errorf("[IntEQ] value isn't equal to %d", i),
	)
}

func IntNE(i int32) ArgOpt {
	return IntOption(
		func(vi int32) bool { return vi != i },
		fmt.Errorf("[IntNE] value isn't not equal to %d", i),
	)
}

func IntLT(i int32) ArgOpt {
	return IntOption(
		func(vi int32) bool { return vi < i },
		fmt.Errorf("[IntLT] value isn't less than %d", i),
	)
}

func IntLTE(i int32) ArgOpt {
	return IntOption(
		func(vi int32) bool { return vi <= i },
		fmt.Errorf("[IntLTE] value isn't less than or equal to %d", i),
	)
}

func IntGT(i int32) ArgOpt {
	return IntOption(
		func(vi int32) bool { return vi > i },
		fmt.Errorf("[IntGT] value isn't greater than %d", i),
	)
}

func IntGTE(i int32) ArgOpt {
	return IntOption(
		func(vi int32) bool { return vi >= i },
		fmt.Errorf("[IntGTE] value isn't greater than or equal to %d", i),
	)
}

func IntPositive() ArgOpt {
	return IntOption(
		func(vi int32) bool { return vi > 0 },
		fmt.Errorf("[IntPositive] value isn't positive"),
	)
}

func IntNonNegative() ArgOpt {
	return IntOption(
		func(vi int32) bool { return vi >= 0 },
		fmt.Errorf("[IntNonNegative] value isn't non-negative"),
	)
}

func IntNegative() ArgOpt {
	return IntOption(
		func(vi int32) bool { return vi < 0 },
		fmt.Errorf("[IntNegative] value isn't negative"),
	)
}

// Float options
func FloatOption(f func(float32) bool, err error) ArgOpt {
	validator := func(v *Value) error {
		if !f(v.GetFloat()) {
			return err
		}
		return nil
	}
	return &option{
		vt:       FloatType,
		validate: validator,
	}
}

func FloatEQ(f float32) ArgOpt {
	return FloatOption(
		func(vf float32) bool { return vf == f },
		fmt.Errorf("[FloatEQ] value isn't equal to %0.2f", f),
	)
}

func FloatNE(f float32) ArgOpt {
	return FloatOption(
		func(vf float32) bool { return vf != f },
		fmt.Errorf("[FloatNE] value isn't not equal to %0.2f", f),
	)
}

func FloatLT(f float32) ArgOpt {
	return FloatOption(
		func(vf float32) bool { return vf < f },
		fmt.Errorf("[FloatLT] value isn't less than %0.2f", f),
	)
}

func FloatLTE(f float32) ArgOpt {
	return FloatOption(
		func(vf float32) bool { return vf <= f },
		fmt.Errorf("[FloatLTE] value isn't less than or equal to %0.2f", f),
	)
}

func FloatGT(f float32) ArgOpt {
	return FloatOption(
		func(vf float32) bool { return vf > f },
		fmt.Errorf("[FloatGT] value isn't greater than %0.2f", f),
	)
}

func FloatGTE(f float32) ArgOpt {
	return FloatOption(
		func(vf float32) bool { return vf >= f },
		fmt.Errorf("[FloatGTE] value isn't greater than or equal to %0.2f", f),
	)
}

func FloatPositive() ArgOpt {
	return FloatOption(
		func(vi float32) bool { return vi > 0 },
		fmt.Errorf("[FloatPositive] value isn't positive"),
	)
}

func FloatNonNegative() ArgOpt {
	return FloatOption(
		func(vi float32) bool { return vi >= 0 },
		fmt.Errorf("[FloatNonNegative] value isn't non-negative"),
	)
}

func FloatNegative() ArgOpt {
	return FloatOption(
		func(vi float32) bool { return vi < 0 },
		fmt.Errorf("[FloatNegative] value isn't negative"),
	)
}

func stringList(s ...string) *Value {
	return &Value{
		Type: &Value_StringList{
			StringList: &StringList{
				List: s,
			},
		},
		Set: true,
	}
}

func intList(l ...int32) *Value {
	return &Value{
		Type: &Value_IntList{
			IntList: &IntList{
				List: l,
			},
		},
		Set: true,
	}
}

func floatList(l ...float32) *Value {
	return &Value{
		Type: &Value_FloatList{
			FloatList: &FloatList{
				List: l,
			},
		},
		Set: true,
	}
}

func boolVal(b bool) *Value {
	return &Value{
		Type: &Value_Bool{
			Bool: b,
		},
		Set: true,
	}
}

func stringVal(s string) *Value {
	return &Value{
		Type: &Value_String_{
			String_: s,
		},
		Set: true,
	}
}

func intVal(i int32) *Value {
	return &Value{
		Type: &Value_Int{
			Int: i,
		},
		Set: true,
	}
}

func floatVal(f float32) *Value {
	return &Value{
		Type: &Value_Float{
			Float: f,
		},
		Set: true,
	}
}
