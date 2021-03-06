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
		if !f(v.String()) {
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
func IntOption(f func(int) bool, err error) ArgOpt {
	validator := func(v *Value) error {
		if !f(v.Int()) {
			return err
		}
		return nil
	}
	return &option{
		vt:       IntType,
		validate: validator,
	}
}

func IntEQ(i int) ArgOpt {
	return IntOption(
		func(vi int) bool { return vi == i },
		fmt.Errorf("[IntEQ] value isn't equal to %d", i),
	)
}

func IntNE(i int) ArgOpt {
	return IntOption(
		func(vi int) bool { return vi != i },
		fmt.Errorf("[IntNE] value isn't not equal to %d", i),
	)
}

func IntLT(i int) ArgOpt {
	return IntOption(
		func(vi int) bool { return vi < i },
		fmt.Errorf("[IntLT] value isn't less than %d", i),
	)
}

func IntLTE(i int) ArgOpt {
	return IntOption(
		func(vi int) bool { return vi <= i },
		fmt.Errorf("[IntLTE] value isn't less than or equal to %d", i),
	)
}

func IntGT(i int) ArgOpt {
	return IntOption(
		func(vi int) bool { return vi > i },
		fmt.Errorf("[IntGT] value isn't greater than %d", i),
	)
}

func IntGTE(i int) ArgOpt {
	return IntOption(
		func(vi int) bool { return vi >= i },
		fmt.Errorf("[IntGTE] value isn't greater than or equal to %d", i),
	)
}

func IntPositive() ArgOpt {
	return IntOption(
		func(vi int) bool { return vi > 0 },
		fmt.Errorf("[IntPositive] value isn't positive"),
	)
}

func IntNonNegative() ArgOpt {
	return IntOption(
		func(vi int) bool { return vi >= 0 },
		fmt.Errorf("[IntNonNegative] value isn't non-negative"),
	)
}

func IntNegative() ArgOpt {
	return IntOption(
		func(vi int) bool { return vi < 0 },
		fmt.Errorf("[IntNegative] value isn't negative"),
	)
}

// Float options
func FloatOption(f func(float64) bool, err error) ArgOpt {
	validator := func(v *Value) error {
		if !f(v.Float()) {
			return err
		}
		return nil
	}
	return &option{
		vt:       FloatType,
		validate: validator,
	}
}

func FloatEQ(f float64) ArgOpt {
	return FloatOption(
		func(vf float64) bool { return vf == f },
		fmt.Errorf("[FloatEQ] value isn't equal to %0.2f", f),
	)
}

func FloatNE(f float64) ArgOpt {
	return FloatOption(
		func(vf float64) bool { return vf != f },
		fmt.Errorf("[FloatNE] value isn't not equal to %0.2f", f),
	)
}

func FloatLT(f float64) ArgOpt {
	return FloatOption(
		func(vf float64) bool { return vf < f },
		fmt.Errorf("[FloatLT] value isn't less than %0.2f", f),
	)
}

func FloatLTE(f float64) ArgOpt {
	return FloatOption(
		func(vf float64) bool { return vf <= f },
		fmt.Errorf("[FloatLTE] value isn't less than or equal to %0.2f", f),
	)
}

func FloatGT(f float64) ArgOpt {
	return FloatOption(
		func(vf float64) bool { return vf > f },
		fmt.Errorf("[FloatGT] value isn't greater than %0.2f", f),
	)
}

func FloatGTE(f float64) ArgOpt {
	return FloatOption(
		func(vf float64) bool { return vf >= f },
		fmt.Errorf("[FloatGTE] value isn't greater than or equal to %0.2f", f),
	)
}

func FloatPositive() ArgOpt {
	return FloatOption(
		func(vi float64) bool { return vi > 0 },
		fmt.Errorf("[FloatPositive] value isn't positive"),
	)
}

func FloatNonNegative() ArgOpt {
	return FloatOption(
		func(vi float64) bool { return vi >= 0 },
		fmt.Errorf("[FloatNonNegative] value isn't non-negative"),
	)
}

func FloatNegative() ArgOpt {
	return FloatOption(
		func(vi float64) bool { return vi < 0 },
		fmt.Errorf("[FloatNegative] value isn't negative"),
	)
}
