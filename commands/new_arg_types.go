package commands

import (
	"fmt"
	"strings"
)

type singleArgProcessor struct {
	optional  bool
	transform func(s string) (*Value, error)
	name      string
	completor *Completor
	vt        ValueType
	// TODO: opts don't need to be here. They can be done in commands.go
	opts []ArgOpt
	// TODO: make separate sub struct for arg vs field values.
	flag      bool
	shortName rune
}

func (sap *singleArgProcessor) set(v *Value, args, flags map[string]*Value) {
	if sap.flag {
		flags[sap.name] = v
	} else {
		args[sap.name] = v
	}
}

func (sap *singleArgProcessor) ProcessExecute(s []string) (*Value, int, error) {
	if len(s) == 0 {
		if sap.optional {
			return nil, 0, nil
		}
		return nil, 0, fmt.Errorf("not enough arguments")
	}
	v, err := sap.transform(s[0])
	return v, 1, err
}

func (sap *singleArgProcessor) ProcessExecuteArgs(rawArgs []string, args, flags map[string]*Value) (int, error) {
	v, n, err := sap.ProcessExecute(rawArgs)
	sap.set(v, args, flags)
	for _, opt := range sap.opts {
		if sap.vt != opt.ValueType() {
			return 0, fmt.Errorf("option can only be bound to arguments with type %v", opt.ValueType())
		}

		if err := opt.Validate(v); err != nil {
			return 0, fmt.Errorf("validation failed: %v", err)
		}
	}
	return n, err
}

func (sap *singleArgProcessor) ProcessCompleteArgs(rawArgs []string, args, flags map[string]*Value) int {
	var v *Value
	var n int
	if len(rawArgs) > 0 {
		v, _ = sap.transform(rawArgs[0])
		n = 1
	}
	sap.set(v, args, flags)
	return n
}

func (sap *singleArgProcessor) Name() string {
	return sap.name
}

func (sap *singleArgProcessor) ShortName() rune {
	return sap.shortName
}

func (sap *singleArgProcessor) Optional() bool {
	return sap.optional
}

func (sap *singleArgProcessor) Complete(rawValue string, args, flags map[string]*Value) *Completion {
	if sap.completor == nil {
		return nil
	}
	var v *Value
	if sap.flag {
		v = flags[sap.name]
	} else {
		v = args[sap.name]
	}
	return sap.completor.Complete(rawValue, v, args, flags)
}

type listArgProcessor struct {
	name      string
	completor *Completor
	opts      []ArgOpt
	minN      int
	optionalN int
	transform func([]string) (*Value, error)
	vt        ValueType
	shortName rune
	flag      bool
}

func (lap *listArgProcessor) set(v *Value, args, flags map[string]*Value) {
	if lap.flag {
		flags[lap.name] = v
	} else {
		args[lap.name] = v
	}
}

func (lap *listArgProcessor) ProcessExecuteArgs(rawArgs []string, args, flags map[string]*Value) (int, error) {
	v, n, err := lap.ProcessExecute(cp(rawArgs))
	lap.set(v, args, flags)
	for _, opt := range lap.opts {
		if lap.vt != opt.ValueType() {
			return 0, fmt.Errorf("option can only be bound to arguments with type %v", opt.ValueType())
		}

		if err := opt.Validate(v); err != nil {
			return 0, fmt.Errorf("validation failed: %v", err)
		}
	}
	return n, err
}

func (lap *listArgProcessor) Name() string {
	return lap.name
}

func (lap *listArgProcessor) ShortName() rune {
	return lap.shortName
}

func (lap *listArgProcessor) Optional() bool {
	return lap.minN == 0
}

func (lap *listArgProcessor) Complete(rawValue string, args, flags map[string]*Value) *Completion {
	if lap.completor == nil {
		return nil
	}
	var v *Value
	if lap.flag {
		v = flags[lap.name]
	} else {
		v = args[lap.name]
	}
	return lap.completor.Complete(rawValue, v, args, flags)
}

func (lap *listArgProcessor) Usage() []string {
	ln := lap.minN
	if lap.optionalN == UnboundedList {
		ln += 1
	} else {
		ln += lap.optionalN
	}

	usage := make([]string, 0, ln)
	n := strings.ReplaceAll(strings.ToUpper(lap.name), " ", "_")
	if lap.flag {
		n = "FLAG_VALUE"
		if lap.shortName == 0 {
			usage = append(usage, fmt.Sprintf("--%s", lap.name))
		} else {
			usage = append(usage, fmt.Sprintf("--%s|-%s", lap.name, string(lap.shortName)))
		}
	}

	for idx := 0; idx < lap.minN; idx++ {
		usage = append(usage, n)
	}

	if lap.optionalN == UnboundedList {
		usage = append(usage, fmt.Sprintf("[%s ...]", n))
	} else if lap.optionalN > 0 {
		usage = append(usage, "[")
		for idx := 0; idx < lap.optionalN; idx++ {
			usage = append(usage, n)
		}
		usage = append(usage, "]")
	}
	return usage
}

func (lap *listArgProcessor) ProcessExecute(s []string) (*Value, int, error) {
	if len(s) < lap.minN {
		return nil, len(s), fmt.Errorf("not enough arguments")
	}
	var endIdx int
	if lap.optionalN == UnboundedList {
		endIdx = len(s)
	} else {
		endIdx = min(lap.minN+lap.optionalN, len(s))
	}
	v, err := lap.transform(s[:endIdx])
	return v, endIdx, err
}

func (lap *listArgProcessor) ProcessCompleteArgs(rawArgs []string, args, flags map[string]*Value) int {
	v, n := lap.ProcessComplete(cp(rawArgs))
	lap.set(v, args, flags)
	/*if lap.flag {
		n = min(n+1, len(rawArgs))
	}*/
	return n
}

func (lap *listArgProcessor) ProcessComplete(s []string) (*Value, int) {
	var endIdx int
	if len(s) < lap.minN || lap.optionalN == UnboundedList {
		endIdx = len(s)
	} else {
		endIdx = min(lap.minN+lap.optionalN, len(s))
	}
	v, _ := lap.transform(s[:endIdx])
	return v, endIdx
}

func (sap *singleArgProcessor) Usage() []string {
	if sap.flag {
		if sap.shortName == 0 {
			return []string{fmt.Sprintf("--%s", sap.name), "FLAG_VALUE"}
		}
		return []string{fmt.Sprintf("--%s|-%s", sap.name, string(sap.shortName)), "FLAG_VALUE"}
	} else if sap.optional {
		return []string{"[", strings.ToUpper(sap.name), "]"}
	}
	return []string{strings.ToUpper(sap.name)}
}

type boolFlagProcessor struct {
	name      string
	shortName rune
}

func (bfp *boolFlagProcessor) ProcessExecuteArgs(rawArgs []string, args, flags map[string]*Value) (int, error) {
	flags[bfp.name] = BoolValue(true)
	return 0, nil
}

func (bfp *boolFlagProcessor) Name() string {
	return bfp.name
}

func (bfp *boolFlagProcessor) ShortName() rune {
	return bfp.shortName
}

func (bfp *boolFlagProcessor) Complete(rawValue string, args, flags map[string]*Value) *Completion {
	return nil
}

func (bfp *boolFlagProcessor) Usage() []string {
	if bfp.shortName != 0 {
		return []string{fmt.Sprintf("--%s|-%s", bfp.name, string(bfp.shortName))}
	}
	return []string{fmt.Sprintf("--%s", bfp.name)}
}

func (bfp *boolFlagProcessor) ProcessCompleteArgs(rawArgs []string, args, flags map[string]*Value) int {
	flags[bfp.name] = BoolValue(true)
	return 0
}
