package commands

import (
	"fmt"
	"strconv"
)

func StringFlag(name string, shortName rune, completor *Completor, opts ...ArgOpt) Flag {
	return &singleArgProcessor{
		name:      name,
		completor: completor,
		opts:      opts,
		vt:        StringType,
		shortName: shortName,
		flag:      true,
		transform: func(s string) (*Value, error) {
			return StringValue(s), nil
		},
	}
}

func IntFlag(name string, shortName rune, completor *Completor, opts ...ArgOpt) Flag {
	return &singleArgProcessor{
		name:      name,
		completor: completor,
		opts:      opts,
		vt:        IntType,
		shortName: shortName,
		flag:      true,
		transform: func(s string) (*Value, error) {
			i, err := strconv.Atoi(s)
			if err != nil {
				err = fmt.Errorf("argument should be an integer: %v", err)
			}
			return IntValue(i), err
		},
	}
}
func FloatFlag(name string, shortName rune, completor *Completor, opts ...ArgOpt) Flag {
	return &singleArgProcessor{
		name:      name,
		completor: completor,
		opts:      opts,
		vt:        IntType,
		shortName: shortName,
		flag:      true,
		transform: func(s string) (*Value, error) {
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				err = fmt.Errorf("argument should be a float: %v", err)
			}
			return FloatValue(f), err
		},
	}
}
func BoolFlag(name string, shortName rune, opts ...ArgOpt) Flag {
	return &boolFlagProcessor{
		name:      name,
		shortName: shortName,
	}
}

func StringListFlag(name string, shortName rune, minN, optionalN int, completor *Completor, opts ...ArgOpt) Flag {
	return &listArgProcessor{
		name:      name,
		minN:      minN,
		optionalN: optionalN,
		completor: completor,
		opts:      opts,
		vt:        StringListType,
		flag:      true,
		shortName: shortName,
		transform: func(s []string) (*Value, error) { return StringListValue(s...), nil },
	}
}

func IntListFlag(name string, shortName rune, minN, optionalN int, completor *Completor, opts ...ArgOpt) Flag {
	return &listArgProcessor{
		name:      name,
		minN:      minN,
		optionalN: optionalN,
		completor: completor,
		opts:      opts,
		vt:        IntListType,
		flag:      true,
		shortName: shortName,
		transform: intListTransform,
	}
}

func FloatListFlag(name string, shortName rune, minN, optionalN int, completor *Completor, opts ...ArgOpt) Flag {
	return &listArgProcessor{
		name:      name,
		minN:      minN,
		optionalN: optionalN,
		completor: completor,
		opts:      opts,
		vt:        FloatListType,
		flag:      true,
		shortName: shortName,
		transform: floatListTransform,
	}
}
