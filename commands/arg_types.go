package commands

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const (
	UnboundedList = -1
)

type ArgOpt interface {
	ValueType() ValueType
	Validate(*Value) error
}

type argProcessor struct {
	ValueType ValueType
	MinN      int
	// Use -1 for unlimited.
	OptionalN int
	argOpts   []ArgOpt
}

func (ap *argProcessor) Value(rawValue []string) (*Value, error) {
	v := &Value{
		valType: ap.ValueType,
	}

	var err error
	switch ap.ValueType {
	case StringType:
		v.stringVal = rawValue[0]
	case StringListType:
		v.stringList = rawValue
	case IntType:
		i, e := strconv.Atoi(rawValue[0])
		if e != nil {
			err = fmt.Errorf("argument should be an integer: %v", e)
		}
		v.intVal = i
	case IntListType:
		var is []int
		for _, rv := range rawValue {
			i, e := strconv.Atoi(rv)
			if e != nil {
				err = fmt.Errorf("int required for IntList argument type: %v", e)
			}
			// TODO: do we want to append the zero value if error or no append at all?
			// TODO: make whatever we decide is reflected in the float.
			// Decided to do this because changing messes up Value.Length function
			is = append(is, i)
		}
		v.intList = is
	case FloatType:
		f, e := strconv.ParseFloat(rawValue[0], 64)
		if e != nil {
			err = fmt.Errorf("argument should be a float: %v", e)
		}
		v.floatVal = f
	case FloatListType:
		var fs []float64
		for _, rv := range rawValue {
			f, e := strconv.ParseFloat(rv, 64)
			if e != nil {
				err = fmt.Errorf("float required for FloatList argument type: %v", e)
			}
			fs = append(fs, f)
		}
		v.floatList = fs
	case BoolType:
		if ap.MinN == 0 && ap.OptionalN == 0 { // flag value, true by presence
			v.boolVal = true
			v.boolFlag = true
		} else { // arg value
			var ok bool
			v.boolVal, ok = boolStringMap[rawValue[0]]
			if !ok {
				var keys []string
				for k := range boolStringMap {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				err = fmt.Errorf("bool value must be one of %v", keys)
			}
		}
	default:
		return nil, fmt.Errorf("invalid value type: %v", ap.ValueType)
	}

	if err != nil {
		return v, err
	}

	for _, opt := range ap.argOpts {
		if ap.ValueType != opt.ValueType() {
			return v, fmt.Errorf("option can only be bound to arguments with type %v", opt.ValueType())
		}

		if err := opt.Validate(v); err != nil {
			return v, fmt.Errorf("validation failed: %v", err)
		}
	}

	return v, nil
}

// TODO: return error instead of boolean
// TODO: should this be ProcessExecuteArgs and ProcessCompleteArgs? feels like weird return values atm.
func (ap *argProcessor) ProcessArgs(args []string) (*Value, bool, error) {
	args, correctNumber := ap.processNumArgs(args)
	value, err := ap.Value(args)
	if err == nil {
		return value, correctNumber, nil
	}
	return value, correctNumber, fmt.Errorf("failed to convert value: %v", err)
}

func (ap *argProcessor) processNumArgs(args []string) ([]string, bool) {
	if ap.UnlimitedN() {
		return args, ap.MinN <= len(args)
	}

	if ap.MinN <= len(args) {
		// Have the minimum number of args, so return the max
		endRange := ap.MinN + ap.OptionalN
		if endRange > len(args) {
			endRange = len(args)
		}
		return args[:endRange], true
	}

	return args, false
}

func (ap *argProcessor) UnlimitedN() bool {
	return ap.OptionalN < 0
}

func (ap *argProcessor) Usage(name string) []string {
	ln := ap.MinN
	if ap.UnlimitedN() {
		ln += 1
	} else {
		ln += ap.OptionalN
	}

	usage := make([]string, 0, ln)
	for idx := 0; idx < ap.MinN; idx++ {
		usage = append(usage, strings.ReplaceAll(strings.ToUpper(name), " ", "_"))
	}

	if ap.UnlimitedN() {
		usage = append(usage, fmt.Sprintf("[%s ...]", strings.ReplaceAll(strings.ToUpper(name), " ", "_")))
	} else if ap.OptionalN > 0 {
		usage = append(usage, "[")
		for idx := 0; idx < ap.OptionalN; idx++ {
			usage = append(usage, strings.ReplaceAll(strings.ToUpper(name), " ", "_"))
		}
		usage = append(usage, "]")
	}
	return usage
}

type genericArgs struct {
	name         string
	argProcessor *argProcessor
	completor    *Completor
}

func (ga *genericArgs) Name() string {
	return ga.name
}

func (ga *genericArgs) Optional() bool {
	return ga.argProcessor.MinN == 0
}

func (ga *genericArgs) Complete(rawValue string, args, flags map[string]*Value) *Completion {
	if ga.completor == nil {
		return nil
	}
	return ga.completor.Complete(rawValue, args[ga.Name()], args, flags)
}

func (ga *genericArgs) ProcessArgs(args []string) (*Value, bool, error) {
	return ga.argProcessor.ProcessArgs(args)
}

func (ga *genericArgs) Usage() []string {
	return ga.argProcessor.Usage(ga.Name())
}

func StringArg(name string, required bool, completor *Completor, opts ...ArgOpt) Arg {
	if !required {
		return listArg(name, StringType, 0, 1, completor, opts...)
	}
	return listArg(name, StringType, 1, 0, completor, opts...)
}

func StringListArg(name string, minN, optionalN int, completor *Completor, opts ...ArgOpt) Arg {
	return listArg(name, StringListType, minN, optionalN, completor, opts...)
}

func IntArg(name string, required bool, completor *Completor, opts ...ArgOpt) Arg {
	if !required {
		return listArg(name, IntType, 0, 1, completor, opts...)
	}
	return listArg(name, IntType, 1, 0, completor, opts...)
}

func IntListArg(name string, minN, optionalN int, completor *Completor, opts ...ArgOpt) Arg {
	return listArg(name, IntListType, minN, optionalN, completor, opts...)
}

func FloatArg(name string, required bool, completor *Completor, opts ...ArgOpt) Arg {
	if !required {
		return listArg(name, FloatType, 0, 1, completor, opts...)
	}
	return listArg(name, FloatType, 1, 0, completor, opts...)
}

func FloatListArg(name string, minN, optionalN int, completor *Completor, opts ...ArgOpt) Arg {
	return listArg(name, FloatListType, minN, optionalN, completor, opts...)
}

func BoolArg(name string, required bool, opts ...ArgOpt) Arg {
	bc := BoolCompletor()
	if required {
		return listArg(name, BoolType, 1, 0, bc, opts...)
	}
	return listArg(name, BoolType, 0, 1, bc, opts...)
}

func listArg(name string, vt ValueType, minN, optionalN int, completor *Completor, opts ...ArgOpt) Arg {
	return &genericArgs{
		name: name,
		argProcessor: &argProcessor{
			MinN:      minN,
			OptionalN: optionalN,
			ValueType: vt,
			argOpts:   opts,
		},
		completor: completor,
	}
}
