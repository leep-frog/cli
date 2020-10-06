package commands

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// Used for testing.
	filepathAbs = filepath.Abs
)

type Completor struct {
	Distinct          bool
	SuggestionFetcher Fetcher
}

type Fetcher interface {
	// Fetch fetches all other options given the command arguments and flags.
	Fetch(value *Value, args, flags map[string]*Value) []string
	PrefixFilter() bool
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
	if c.SuggestionFetcher.PrefixFilter() {
		for _, o := range allOpts {
			if strings.HasPrefix(o, lastArg) {
				filteredOpts = append(filteredOpts, o)
			}
		}
	} else {
		filteredOpts = allOpts
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
func (nf *NoopFetcher) PrefixFilter() bool                              { return true }

type ListFetcher struct {
	Options []string
}

func (lf *ListFetcher) Fetch(_ *Value, _, _ map[string]*Value) []string { return lf.Options }
func (lf *ListFetcher) PrefixFilter() bool                              { return true }

// TODO: this needs to complete the second half of the command as well
type FileFetcher struct {
	Regexp            *regexp.Regexp
	Directory         string
	IgnoreFiles       bool
	IgnoreDirectories bool
}

func (ff *FileFetcher) PrefixFilter() bool { return false }

// TODO: should these be allowed to return errors?
func (ff *FileFetcher) Fetch(value *Value, args, flags map[string]*Value) []string {
	var lastArg string
	if strPtr := value.String(); strPtr != nil {
		lastArg = *strPtr
	} else if slPtr := value.StringList(); slPtr != nil && len(*slPtr) > 0 {
		lastArg = (*slPtr)[len(*slPtr)-1]
	}

	fmt.Printf("LA_%s_%s\n", lastArg, ff.Directory)

	laDir, laFile := filepath.Split(lastArg)
	dir, err := filepathAbs(filepath.Join(ff.Directory, laDir))
	if err != nil {
		return nil
	}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil
	}

	onlyDir := true
	suggestions := make([]string, 0, len(files))
	for _, f := range files {
		if (f.Mode().IsDir() && ff.IgnoreDirectories) || (f.Mode().IsRegular() && ff.IgnoreFiles) {
			continue
		}

		if ff.Regexp != nil && !ff.Regexp.MatchString(f.Name()) {
			continue
		}

		if !strings.HasPrefix(f.Name(), laFile) {
			continue
		}

		if f.Mode().IsDir() {
			suggestions = append(suggestions, fmt.Sprintf("%s/", f.Name()))
		} else {
			onlyDir = false
			suggestions = append(suggestions, f.Name())
		}
	}

	// If only 1 suggestion matching, then we want it to autocomplete the whole thing.
	if len(suggestions) == 1 {
		// Want to autocomplete the full path
		// Note: we can't use filepath.Join here because it cleans up the path
		suggestions[0] = fmt.Sprintf("%s%s", laDir, suggestions[0])

		if onlyDir {
			// This does dir1/ and dir1// so that the user's command is autocompleted to dir1/
			// without a space after it.
			suggestions = append(suggestions, fmt.Sprintf("%s/", suggestions[0]))
		}
	}
	return suggestions
}

// TODO type MultiFetcher struct { cs Fetchers }
