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
	opts      []ArgOpt
}

func (sap *singleArgProcessor) ProcessExecuteArgs(rawArgs []string, args, flags map[string]*Value) ([]string, bool, error) {
	v, n, err := sap.ProcessExecute(rawArgs)
	args[sap.name] = v
	for _, opt := range sap.opts {
		if sap.vt != opt.ValueType() {
			return nil, false, fmt.Errorf("option can only be bound to arguments with type %v", opt.ValueType())
		}

		if err := opt.Validate(v); err != nil {
			return nil, false, fmt.Errorf("validation failed: %v", err)
		}
	}
	return rawArgs[n:], false, err
}

func (sap *singleArgProcessor) ProcessCompleteArgs(rawArgs []string, args, flags map[string]*Value) int {
	v, n := sap.ProcessComplete(rawArgs)
	args[sap.name] = v
	return n
}

func (sap *singleArgProcessor) Name() string {
	return sap.name
}

func (sap *singleArgProcessor) Optional() bool {
	return sap.optional
}

func (sap *singleArgProcessor) Complete(rawValue string, args, flags map[string]*Value) *Completion {
	if sap.completor == nil {
		return nil
	}
	return sap.completor.Complete(rawValue, args[sap.Name()], args, flags)
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

func (sap *singleArgProcessor) ProcessComplete(s []string) (*Value, int) {
	if len(s) == 0 {
		return nil, 0
	}
	v, _ := sap.transform(s[0])
	return v, 1
}

type listArgProcessor struct {
	name      string
	completor *Completor
	opts      []ArgOpt
	minN      int
	optionalN int
	transform func([]string) (*Value, error)
	vt        ValueType
}

func (lap *listArgProcessor) ProcessExecuteArgs(rawArgs []string, args, flags map[string]*Value) ([]string, bool, error) {
	v, n, err := lap.ProcessExecute(rawArgs)
	args[lap.name] = v
	for _, opt := range lap.opts {
		if lap.vt != opt.ValueType() {
			return nil, false, fmt.Errorf("option can only be bound to arguments with type %v", opt.ValueType())
		}

		if err := opt.Validate(v); err != nil {
			return nil, false, fmt.Errorf("validation failed: %v", err)
		}
	}
	return rawArgs[n:], false, err
}

func (lap *listArgProcessor) ProcessCompleteArgs(rawArgs []string, args, flags map[string]*Value) int {
	v, n := lap.ProcessComplete(cp(rawArgs))
	args[lap.name] = v
	return n
}

func (lap *listArgProcessor) Name() string {
	return lap.name
}

func (lap *listArgProcessor) Optional() bool {
	return lap.minN == 0
}

func (lap *listArgProcessor) Complete(rawValue string, args, flags map[string]*Value) *Completion {
	if lap.completor == nil {
		return nil
	}
	return lap.completor.Complete(rawValue, args[lap.Name()], args, flags)
}

func (lap *listArgProcessor) Usage() []string {
	ln := lap.minN
	if lap.optionalN == UnboundedList {
		ln += 1
	} else {
		ln += lap.optionalN
	}

	usage := make([]string, 0, ln)
	for idx := 0; idx < lap.minN; idx++ {
		usage = append(usage, strings.ReplaceAll(strings.ToUpper(lap.name), " ", "_"))
	}

	if lap.optionalN == UnboundedList {
		usage = append(usage, fmt.Sprintf("[%s ...]", strings.ReplaceAll(strings.ToUpper(lap.name), " ", "_")))
	} else if lap.optionalN > 0 {
		usage = append(usage, "[")
		for idx := 0; idx < lap.optionalN; idx++ {
			usage = append(usage, strings.ReplaceAll(strings.ToUpper(lap.name), " ", "_"))
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
	if sap.optional {
		return []string{"[", strings.ToUpper(sap.name), "]"}
	}
	return []string{strings.ToUpper(sap.name)}
}
