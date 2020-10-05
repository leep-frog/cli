package commands

import (
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

var (
	// Getwd gets the current working directory (needed to stub out in tests).
	getwd = os.Getwd
)

type Completor struct {
	Distinct          bool
	SuggestionFetcher Fetcher
}

type Fetcher interface {
	// Fetch fetches all other options given the command arguments and flags.
	Fetch(value *Value, args, flags map[string]*Value) []string
}

// TODO: values arg should be a *Value
func (c *Completor) Complete(value *Value, args, flags map[string]*Value) []string {
	if c == nil || c.SuggestionFetcher == nil {
		return nil
	}

	allOpts := c.SuggestionFetcher.Fetch(value, args, flags)

	// Filter out prefixes (this should be optional based on Completor.FilterPrefix (or option?))
	var lastArg string
	if strPtr := value.String(); strPtr != nil {
		lastArg = *strPtr
	} else if slPtr := value.StringList(); slPtr != nil && len(*slPtr) > 0 {
		lastArg = (*slPtr)[len(*slPtr)-1]
	}

	var filteredOpts []string
	for _, o := range allOpts {
		if strings.HasPrefix(o, lastArg) {
			filteredOpts = append(filteredOpts, o)
		}
	}

	if !c.Distinct || value.valType != StringListType {
		// TODO: if we ever want to autocomplete non-string types, we should make Fetch
		// return Value types (and add public methods to construct int, string, float values).
		return filteredOpts
	}

	existingValues := map[string]bool{}
	for _, s := range *value.StringList() {
		existingValues[s] = true
	}

	var distinctOpts []string
	for _, opt := range filteredOpts {
		if !existingValues[opt] {
			distinctOpts = append(distinctOpts, opt)
		}
	}
	return distinctOpts
}

type NoopFetcher struct{}

func (nf *NoopFetcher) Fetch(_ *Value, _, _ map[string]*Value) []string { return nil }

type ListFetcher struct {
	Options []string
}

func (lf *ListFetcher) Fetch(_ *Value, _, _ map[string]*Value) []string { return lf.Options }

// TODO: this needs to complete the second half of the command as well
type FileFetcher struct {
	Regexp            *regexp.Regexp
	Directory         string
	IgnoreFiles       bool
	IgnoreDirectories bool
}

// TODO: should these be allowed to return errors?
func (ff *FileFetcher) Fetch(value *Value, args, flags map[string]*Value) []string {
	dir := ff.Directory
	if dir == "" {
		var err error
		dir, err = getwd()
		if err != nil {
			return nil
		}
	}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil
	}

	suggestions := make([]string, 0, len(files))
	for _, f := range files {
		if (f.Mode().IsDir() && ff.IgnoreDirectories) || (f.Mode().IsRegular() && ff.IgnoreFiles) {
			continue
		}

		if ff.Regexp != nil && !ff.Regexp.MatchString(f.Name()) {
			continue
		}

		suggestions = append(suggestions, f.Name())
	}

	return suggestions
}

// TODO type MultiFetcher struct { cs Fetchers }
