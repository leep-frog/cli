package commands

import (
	"fmt"
	"strconv"
)

const (
	UnboundedList = -1
)

type ArgOpt interface {
	ValueType() ValueType
	Validate(*Value) error
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func cp(s []string) []string {
	r := make([]string, 0, len(s))
	for _, i := range s {
		r = append(r, i)
	}
	return r
}

func StringArg(name string, required bool, completor *Completor, opts ...ArgOpt) Arg {
	return &singleArgProcessor{
		name:      name,
		completor: completor,
		opts:      opts,
		vt:        StringType,
		optional:  !required,
		transform: func(s string) (*Value, error) {
			return StringValue(s), nil
		},
	}
}

func IntArg(name string, required bool, completor *Completor, opts ...ArgOpt) Arg {
	return &singleArgProcessor{
		name:      name,
		optional:  !required,
		completor: completor,
		opts:      opts,
		vt:        IntType,
		transform: func(s string) (*Value, error) {
			i, err := strconv.Atoi(s)
			if err != nil {
				err = fmt.Errorf("argument should be an integer: %v", err)
			}
			return IntValue(i), err
		},
	}
}

func FloatArg(name string, required bool, completor *Completor, opts ...ArgOpt) Arg {
	return &singleArgProcessor{
		name:      name,
		completor: completor,
		opts:      opts,
		vt:        FloatType,
		optional:  !required,
		transform: func(s string) (*Value, error) {
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				err = fmt.Errorf("argument should be a float: %v", err)
			}
			return FloatValue(f), err
		},
	}
}

func BoolArg(name string, required bool, opts ...ArgOpt) Arg {
	return &singleArgProcessor{
		name:      name,
		completor: BoolCompletor(),
		opts:      opts,
		vt:        FloatType,
		optional:  !required,
		transform: func(s string) (*Value, error) {
			b, err := strconv.ParseBool(s)
			if err != nil {
				err = fmt.Errorf("argument should be a bool: %v", err)
			}
			return BoolValue(b), err
		},
	}
}

func StringListArg(name string, minN, optionalN int, completor *Completor, opts ...ArgOpt) Arg {
	return &listArgProcessor{
		name:      name,
		minN:      minN,
		optionalN: optionalN,
		completor: completor,
		opts:      opts,
		vt:        StringListType,
		transform: func(s []string) (*Value, error) { return StringListValue(s...), nil },
	}
}

func IntListArg(name string, minN, optionalN int, completor *Completor, opts ...ArgOpt) Arg {
	return &listArgProcessor{
		name:      name,
		minN:      minN,
		optionalN: optionalN,
		completor: completor,
		opts:      opts,
		vt:        IntListType,
		transform: intListTransform,
	}
}

func intListTransform(sl []string) (*Value, error) {
	var err error
	var is []int
	for _, s := range sl {
		i, e := strconv.Atoi(s)
		if e != nil {
			// TODO: add failed to load field to values.
			// These can be used in autocomplete if necessary.
			err = e
		}
		is = append(is, i)
	}
	return IntListValue(is...), err
}

func FloatListArg(name string, minN, optionalN int, completor *Completor, opts ...ArgOpt) Arg {
	return &listArgProcessor{
		name:      name,
		minN:      minN,
		optionalN: optionalN,
		completor: completor,
		opts:      opts,
		vt:        FloatListType,
		transform: floatListTransform,
	}
}

func floatListTransform(sl []string) (*Value, error) {
	var err error
	var fs []float64
	for _, s := range sl {
		f, e := strconv.ParseFloat(s, 64)
		if e != nil {
			err = e
		}
		fs = append(fs, f)
	}
	return FloatListValue(fs...), err
}
