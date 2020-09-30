package commands

import "fmt"

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

func (gf *genericFlag) Complete(args, flags map[string]*Value) []string {
	if gf.completor == nil {
		return nil
	}
	return gf.completor.Complete(flags[gf.Name()], args, flags)
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

func (gf *genericFlag) ProcessArgs(args []string) (*Value, bool, error) {
	return gf.argProcessor.ProcessArgs(args)
}

// TODO: this
// NewBooleanFlag returns a boolean flag
func NewBooleanFlag(name string, shortName rune, default_ bool) Flag {
	return &genericFlag{
		name:         name,
		shortName:    shortName,
		argProcessor: &argProcessor{},
	}
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

func listFlag(name string, shortName rune, vt ValueType, minN, optionalN int, completor *Completor, opts ...ArgOpt) Flag {
	return &genericFlag{
		name:      name,
		shortName: shortName,
		argProcessor: &argProcessor{
			MinN:      minN,
			OptionalN: optionalN,
			ValueType: vt,
			argOpts:   opts,
		},
		completor: completor,
	}
}
