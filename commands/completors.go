package commands

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

var (
	// Used for testing.
	filepathAbs = filepath.Abs
)

const (
	suffixChar = "_"
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

	// Don't include the last element because sometimes we want to just add a
	// a space to the command. For example,
	// "e commands.go" should return ["commands.go"]
	sl := *value.StringList()
	for i := 0; i < len(sl)-1; i++ {
		existingValues[sl[i]] = true
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
	Regexp    *regexp.Regexp
	Directory string
	// Whether or not each argument has to be unique.
	// Separate from Completor.Distinct because file fetching
	// does more complicated custom logic.
	Distinct          bool
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

	// Remove any non-distinct matches, if relevant.
	if ff.Distinct {
		valSet := map[string]bool{}
		// TODO: make validation function so file fetchers can only be
		// assigned to string lists.
		for _, v := range *value.StringList() {
			valSet[v] = true
		}

		distinctSuggestions := make([]string, 0, len(suggestions))
		for _, s := range suggestions {
			if !valSet[fmt.Sprintf("%s%s", laDir, s)] {
				distinctSuggestions = append(distinctSuggestions, s)
			}
		}
		if len(distinctSuggestions) == 0 {
			return nil
		}
		suggestions = distinctSuggestions
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
			c.Suggestions = append(c.Suggestions, fmt.Sprintf("%s%s", c.Suggestions[0], suffixChar))
		}
		return c
	}

	autoFill, ok := getAutofillLetters(laFile, c.Suggestions)
	if !ok {
		// Nothing can be autofilled so we just return file names
		// Don't autocomplete because all suggestions have the same
		// prefix so this would actually autocomplete to the prefix
		// without the directory name
		c.DontComplete = true
		return c
	}

	// Otherwise, we should complete all of the autofill letters
	c.DontComplete = false
	autoFill = laDir + autoFill
	c.Suggestions = []string{
		autoFill,
		autoFill + suffixChar,
	}
	return c
}

func getAutofillLetters(laFile string, suggestions []string) (string, bool) {
	nextLetterPos := len(laFile)
	for proceed := true; proceed; nextLetterPos++ {
		var nextLetter *rune
		var lowerNextLetter rune
		for _, s := range suggestions {
			if len(s) <= nextLetterPos {
				// If a remaining suggestion has run out of letters, then
				// we can't autocomplete more than that.
				proceed = false
				break
			}

			char := rune(s[nextLetterPos])
			if nextLetter == nil {
				nextLetter = &char
				lowerNextLetter = unicode.ToLower(char)
				continue
			}

			if unicode.ToLower(char) != lowerNextLetter {
				proceed = false
				break
			}
		}
	}

	completeUpTo := nextLetterPos - 1
	if completeUpTo <= len(laFile) {
		return "", false
	}

	caseToCompleteWith := suggestions[0]
	for _, s := range suggestions {
		if strings.HasPrefix(s, laFile) {
			caseToCompleteWith = s
			break
		}
	}
	return caseToCompleteWith[:completeUpTo], true
}
