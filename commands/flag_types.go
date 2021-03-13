package commands

import (
	"fmt"
	"strconv"
)

type genericFlag struct {
	name         string
	shortName    rune
	argProcessor *argProcessor
	completor    *Completor
}

func (gf *genericFlag) Name() string {
	return gf.name
}

func (gf *genericFlag) ShortName() rune {
	return gf.shortName
}

func (gf *genericFlag) Complete(rawValue string, args, flags map[string]*Value) *Completion {
	if gf.completor == nil {
		return nil
	}
	return gf.completor.Complete(rawValue, flags[gf.Name()], args, flags)
}

func (gf *genericFlag) Length(v *Value) int {
	if v.IsType(BoolType) {
		return 0
	}
	return v.Length()
}

func (gf *genericFlag) Usage() []string {
	var flagString string
	if gf.ShortName() == 0 {
		flagString = fmt.Sprintf("--%s", gf.Name())
	} else {
		flagString = fmt.Sprintf("--%s|-%s", gf.Name(), string(gf.ShortName()))
	}
	// TODO: better name than FLAG_VALUE.
	return append([]string{flagString}, gf.argProcessor.Usage("FLAG_VALUE")...)
}

func (gf *genericFlag) ProcessCompleteArgs(rawArgs []string, args, flags map[string]*Value) int {
	return gf.argProcessor.ProcessCompleteArgs(rawArgs, args, flags)
}

func (gf *genericFlag) ProcessExecuteArgs(rawArgs []string, args, flags map[string]*Value) ([]string, bool, error) {
	return gf.argProcessor.ProcessExecuteArgs(rawArgs, args, flags)
}

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
	return &genericFlag{
		name:      name,
		shortName: shortName,
		argProcessor: &argProcessor{
			ValueType: BoolType,
			argOpts:   opts,
			flag:      true,
			argName:   name,
			boolFlag:  true,
		},
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
