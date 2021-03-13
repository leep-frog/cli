package commands

import (
	"fmt"
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
	return listFlag(name, shortName, StringType, 1, 0, completor, opts...)
}

func StringListFlag(name string, shortName rune, minN, optionalN int, completor *Completor, opts ...ArgOpt) Flag {
	return listFlag(name, shortName, StringListType, minN, optionalN, completor, opts...)
}

func IntFlag(name string, shortName rune, completor *Completor, opts ...ArgOpt) Flag {
	return listFlag(name, shortName, IntType, 1, 0, completor, opts...)
}

func IntListFlag(name string, shortName rune, minN, optionalN int, completor *Completor, opts ...ArgOpt) Flag {
	return listFlag(name, shortName, IntListType, minN, optionalN, completor, opts...)
}

func FloatFlag(name string, shortName rune, completor *Completor, opts ...ArgOpt) Flag {
	return listFlag(name, shortName, FloatType, 1, 0, completor, opts...)
}

func FloatListFlag(name string, shortName rune, minN, optionalN int, completor *Completor, opts ...ArgOpt) Flag {
	return listFlag(name, shortName, FloatListType, minN, optionalN, completor, opts...)
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

func listFlag(name string, shortName rune, vt ValueType, minN, optionalN int, completor *Completor, opts ...ArgOpt) Flag {
	return &genericFlag{
		name:      name,
		shortName: shortName,
		argProcessor: &argProcessor{
			MinN:      minN,
			OptionalN: optionalN,
			ValueType: vt,
			argOpts:   opts,
			flag:      true,
			argName:   name,
		},
		completor: completor,
	}
}
