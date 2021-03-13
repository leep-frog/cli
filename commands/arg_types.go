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

type argTypeProcessor interface {
	ProcessExecute([]string) (*Value, int, error)
	ProcessComplete([]string) (*Value, int)
}

type argProcessor struct {
	ValueType ValueType
	MinN      int
	// Use -1 for unlimited.
	OptionalN        int
	argOpts          []ArgOpt
	flag             bool
	argName          string
	argTypeProcessor argTypeProcessor
	boolFlag         bool
}

func (ap *argProcessor) Value(rawValue []string) (*Value, error) {
	var v *Value

	var err error
	switch ap.ValueType {
	case StringType:
		v = StringValue(rawValue[0])
	case StringListType:
		v = StringListValue(rawValue...)
	case IntType:
		i, e := strconv.Atoi(rawValue[0])
		if e != nil {
			err = fmt.Errorf("argument should be an integer: %v", e)
		}
		v = IntValue(int(i))
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
			is = append(is, int(i))
		}
		v = IntListValue(is...)
	case FloatType:
		f, e := strconv.ParseFloat(rawValue[0], 64)
		if e != nil {
			err = fmt.Errorf("argument should be a float: %v", e)
		}

		v = FloatValue(float64(f))
	case FloatListType:
		var fs []float64
		for _, rv := range rawValue {
			f, e := strconv.ParseFloat(rv, 64)
			if e != nil {
				err = fmt.Errorf("float required for FloatList argument type: %v", e)
			}
			fs = append(fs, float64(f))
		}
		v = FloatListValue(fs...)
	case BoolType:
		if ap.MinN == 0 && ap.OptionalN == 0 { // flag value, true by presence
			v = BoolValue(true)
		} else { // arg value
			var b, ok bool
			b, ok = boolStringMap[rawValue[0]]
			if !ok {
				var keys []string
				for k := range boolStringMap {
					keys = append(keys, k)
				}
				sort.Strings(keys)
				err = fmt.Errorf("bool value must be one of %v", keys)
			}
			v = BoolValue(b)
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

func (ap *argProcessor) Set(v *Value, args, flags map[string]*Value) {
	if ap.flag {
		flags[ap.argName] = v
	} else {
		args[ap.argName] = v
	}
}

// TODO: return error instead of boolean
func (ap *argProcessor) ProcessCompleteArgs(rawArgs []string, args, flags map[string]*Value) int {
	if ap.argTypeProcessor != nil {
		v, n := ap.argTypeProcessor.ProcessComplete(cp(rawArgs))
		ap.Set(v, args, flags)
		return n
	}
	newArgs, _ := ap.processNumArgs(rawArgs)
	value, err := ap.Value(cp(newArgs))
	ap.Set(value, args, flags)
	if err == nil {
		if ap.boolFlag {
			return 0
		}
		return value.Length()
	}

	var n int
	if ap.UnlimitedN() {
		n = len(rawArgs)
	} else {
		n = min(ap.MinN+ap.OptionalN, len(rawArgs))
	}
	// Don't return error because we still want to process.
	return n //fmt.Errorf("failed to convert value: %v", err)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
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

func (ap *argProcessor) ProcessExecuteArgs(rawArgs []string, args, flags map[string]*Value) ([]string, bool, error) {
	if ap.argTypeProcessor != nil {
		v, n, err := ap.argTypeProcessor.ProcessExecute(rawArgs)
		if err != nil {
			return nil, false, fmt.Errorf("failed to process %q arg: %v", ap.argName, err)
		}
		ap.Set(v, args, flags)

		// TODO: move this somewhere else.
		for _, opt := range ap.argOpts {
			if ap.ValueType != opt.ValueType() {
				return nil, false, fmt.Errorf("option can only be bound to arguments with type %v", opt.ValueType())
			}

			if err := opt.Validate(v); err != nil {
				return nil, false, fmt.Errorf("validation failed: %v", err)
			}
		}

		return rawArgs[n:], true, nil
	}
	argsForValue, correctNumber := ap.processNumArgs(rawArgs)
	value, err := ap.Value(cp(argsForValue))
	ap.Set(value, args, flags)
	if err != nil {
		return nil, correctNumber, fmt.Errorf("failed to convert value: %v", err)
	}
	if !correctNumber {
		return nil, false, fmt.Errorf("not enough arguments for %q arg", ap.argName)
	}
	return rawArgs[len(argsForValue):], correctNumber, nil
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

func (ga *genericArgs) ProcessCompleteArgs(rawArgs []string, args, flags map[string]*Value) int {
	return ga.argProcessor.ProcessCompleteArgs(rawArgs, args, flags)
}

func (ga *genericArgs) ProcessExecuteArgs(rawArgs []string, args, flags map[string]*Value) ([]string, bool, error) {
	return ga.argProcessor.ProcessExecuteArgs(rawArgs, args, flags)
}

func (ga *genericArgs) Usage() []string {
	return ga.argProcessor.Usage(ga.Name())
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
			if b, ok := boolStringMap[s]; ok {
				return BoolValue(b), nil
			}

			var keys []string
			for k := range boolStringMap {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			return nil, fmt.Errorf("bool value must be one of %v", keys)
		},
	}
}

func StringListArg(name string, minN, optionalN int, completor *Completor, opts ...ArgOpt) Arg {
	p := &listArgProcessor{
		minN:      minN,
		optionalN: optionalN,
		transform: func(s []string) (*Value, error) { return StringListValue(s...), nil },
	}
	return newListArg(name, StringListType, minN, optionalN, completor, p, opts...)
}

func IntListArg(name string, minN, optionalN int, completor *Completor, opts ...ArgOpt) Arg {
	p := &listArgProcessor{
		minN:      minN,
		optionalN: optionalN,
		transform: func(sl []string) (*Value, error) {
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
		},
	}
	return newListArg(name, IntListType, minN, optionalN, completor, p, opts...)
}

func FloatListArg(name string, minN, optionalN int, completor *Completor, opts ...ArgOpt) Arg {
	p := &listArgProcessor{
		minN:      minN,
		optionalN: optionalN,
		transform: func(sl []string) (*Value, error) {
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
		},
	}
	return newListArg(name, IntListType, minN, optionalN, completor, p, opts...)
}

func newListArg(name string, vt ValueType, minN, optionalN int, completor *Completor, processor argTypeProcessor, opts ...ArgOpt) Arg {
	return &genericArgs{
		name: name,
		argProcessor: &argProcessor{
			MinN:             minN,
			OptionalN:        optionalN,
			ValueType:        vt,
			argOpts:          opts,
			argName:          name,
			argTypeProcessor: processor,
		},
		completor: completor,
	}
}
