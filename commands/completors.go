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

type Completion struct {
	Suggestions        []string
	IgnoreFilter       bool
	DontComplete       bool
	CaseInsenstiveSort bool
}

func BoolCompletor() *Completor {
	return &Completor{
		SuggestionFetcher: &boolFetcher{},
	}
}

type boolFetcher struct{}

func (*boolFetcher) Fetch(value *Value, args, flags map[string]*Value) *Completion {
	var keys []string
	for k := range boolStringMap {
		keys = append(keys, k)
	}
	return &Completion{
		Suggestions: keys,
	}
}

type Fetcher interface {
	// Fetch fetches all other options given the command arguments and flags.
	Fetch(value *Value, args, flags map[string]*Value) *Completion
}

func (c *Completor) Complete(rawValue string, value *Value, args, flags map[string]*Value) *Completion {
	if c == nil || c.SuggestionFetcher == nil {
		return nil
	}

	completion := c.SuggestionFetcher.Fetch(value, args, flags)
	if completion == nil {
		return nil
	}
	allOpts := completion.Suggestions

	// Filter out prefixes.
	if !completion.IgnoreFilter {
		var filteredOpts []string
		for _, o := range allOpts {
			if strings.HasPrefix(o, rawValue) {
				filteredOpts = append(filteredOpts, o)
			}
		}
		completion.Suggestions = filteredOpts
	}

	if !c.Distinct || value.valType != StringListType {
		// TODO: if we ever want to autocomplete non-string types, we should make Fetch
		// return Value types (and add public methods to construct int, string, float values).
		return completion
	}

	existingValues := map[string]bool{}
	for _, s := range *value.StringList() {
		existingValues[s] = true
	}

	var distinctOpts []string
	for _, opt := range completion.Suggestions {
		if !existingValues[opt] {
			distinctOpts = append(distinctOpts, opt)
		}
	}
	completion.Suggestions = distinctOpts
	return completion
}

type NoopFetcher struct{}

func (nf *NoopFetcher) Fetch(_ *Value, _, _ map[string]*Value) *Completion { return nil }

type ListFetcher struct {
	Options []string
}

func (lf *ListFetcher) Fetch(_ *Value, _, _ map[string]*Value) *Completion {
	return &Completion{Suggestions: lf.Options}
}

type FileFetcher struct {
	Regexp            *regexp.Regexp
	Directory         string
	IgnoreFiles       bool
	IgnoreDirectories bool
}

func (ff *FileFetcher) Fetch(value *Value, args, flags map[string]*Value) *Completion {
	var lastArg string
	if strPtr := value.String(); strPtr != nil {
		lastArg = *strPtr
	} else if slPtr := value.StringList(); slPtr != nil && len(*slPtr) > 0 {
		lastArg = (*slPtr)[len(*slPtr)-1]
	}

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

		if !strings.HasPrefix(strings.ToLower(f.Name()), strings.ToLower(laFile)) {
			continue
		}

		if f.Mode().IsDir() {
			suggestions = append(suggestions, fmt.Sprintf("%s/", f.Name()))
		} else {
			onlyDir = false
			suggestions = append(suggestions, f.Name())
		}
	}

	if len(suggestions) == 0 {
		return nil
	}

	c := &Completion{
		Suggestions:        suggestions,
		IgnoreFilter:       true,
		CaseInsenstiveSort: true,
	}

	// If only 1 suggestion matching, then we want it to autocomplete the whole thing.
	if len(c.Suggestions) == 1 {
		// Want to autocomplete the full path
		// Note: we can't use filepath.Join here because it cleans up the path
		c.Suggestions[0] = fmt.Sprintf("%s%s", laDir, c.Suggestions[0])

		if onlyDir {
			// This does dir1/ and dir1// so that the user's command is autocompleted to dir1/
			// without a space after it.
			c.Suggestions = append(c.Suggestions, fmt.Sprintf("%s/", c.Suggestions[0]))
		}
		return c
	}

	// We check if all suggestions match laFile, because then we can
	// autofill letters that are the same for all options.
	var nextLetter *rune
	nextLetterPos := len(laFile)
	for _, s := range c.Suggestions {
		if len(s) > nextLetterPos {
			if nextLetter == nil {
				rn := rune(s[nextLetterPos])
				nextLetter = &rn
			} else if rune(s[nextLetterPos]) != *nextLetter {
				// If two options differ in next letter, then no extra letters can be
				// filled. However, this is only a problem if we are completing an
				// argument for a sub-directory.
				c.DontComplete = len(laDir) != 0
				return c
			}
		}
	}

	// If we are here, then we can autofill some letters
	for i, s := range c.Suggestions {
		c.Suggestions[i] = fmt.Sprintf("%s%s", laDir, s)
	}
	return c
}
