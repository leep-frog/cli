package commands

import (
	"fmt"
	"sort"
	"strconv"
)

type stringArgProcessor struct {
	optional bool
}

func (sap *stringArgProcessor) ProcessExecute(s []string) (*Value, int, error) {
	if len(s) == 0 {
		if sap.optional {
			return nil, 0, nil
		}
		return nil, 0, fmt.Errorf("not enough arguments")
	}
	return StringValue(s[0]), 1, nil
}

func (sap *stringArgProcessor) ProcessComplete(s []string) (*Value, int) {
	if len(s) == 0 {
		return nil, 0
	}
	return StringValue(s[0]), 1
}

type listArgProcessor struct {
	minN      int
	optionalN int
	transform func([]string) (*Value, error)
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

type intArgProcessor struct {
	optional bool
}

func (iap *intArgProcessor) ProcessExecute(s []string) (*Value, int, error) {
	if len(s) == 0 {
		if iap.optional {
			return nil, 0, nil
		}
		return nil, 0, fmt.Errorf("not enough arguments")
	}
	i, err := strconv.Atoi(s[0])
	if err != nil {
		return nil, 1, fmt.Errorf("argument should be an integer: %v", err)
	}
	return IntValue(i), 1, nil
}

func (iap *intArgProcessor) ProcessComplete(s []string) (*Value, int) {
	if len(s) == 0 {
		return nil, 0
	}
	i, err := strconv.Atoi(s[0])
	if err != nil {
		return IntValue(0), 1
	}
	return IntValue(i), 1
}

type floatArgProcessor struct {
	optional bool
}

func (fap *floatArgProcessor) ProcessExecute(s []string) (*Value, int, error) {
	if len(s) == 0 {
		if fap.optional {
			return nil, 0, nil
		}
		return nil, 0, fmt.Errorf("not enough arguments")
	}
	f, err := strconv.ParseFloat(s[0], 64)
	if err != nil {
		return nil, 1, fmt.Errorf("argument should be a float: %v", err)
	}
	return FloatValue(f), 1, nil
}

func (fap *floatArgProcessor) ProcessComplete(s []string) (*Value, int) {
	if len(s) == 0 {
		return nil, 0
	}
	f, err := strconv.ParseFloat(s[0], 64)
	if err != nil {
		return FloatValue(0), 1
	}
	return FloatValue(f), 1
}

type boolArgProcessor struct {
	optional bool
}

func (bap *boolArgProcessor) ProcessExecute(s []string) (*Value, int, error) {
	if len(s) == 0 {
		if bap.optional {
			return nil, 0, nil
		}
		return nil, 0, fmt.Errorf("not enough arguments")
	}
	if b, ok := boolStringMap[s[0]]; ok {
		return BoolValue(b), 1, nil
	}

	var keys []string
	for k := range boolStringMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return nil, 1, fmt.Errorf("bool value must be one of %v", keys)
}

func (bap *boolArgProcessor) ProcessComplete(s []string) (*Value, int) {
	if len(s) == 0 {
		return nil, 0
	}
	f, err := strconv.ParseFloat(s[0], 64)
	if err != nil {
		return FloatValue(0), 1
	}
	return FloatValue(f), 1
}
