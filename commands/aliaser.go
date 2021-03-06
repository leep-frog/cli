package commands

import (
	"fmt"
	"regexp"
	"sort"
)

const (
	AliasArg  = "ALIAS"
	RegexpArg = "REGEXP"
)

type AliasVerifier func(cos CommandOS, alias string, value *Value, args, flags map[string]*Value) bool

type AliasTransformer func(cos CommandOS, alias string, value *Value, args, flags map[string]*Value) (*Value, bool)

type Aliaser struct {
	Arg         Arg
	Verifier    AliasVerifier
	Transformer AliasTransformer

	Aliases map[string]*Value

	changed bool
}

func (a *Aliaser) Changed() bool {
	return a.changed
}

type aliasFetcher struct {
	a *Aliaser
}

func (af *aliasFetcher) Fetch(value *Value, args, flags map[string]*Value) *Completion {
	suggestions := make([]string, 0, len(af.a.Aliases))
	for k := range af.a.Aliases {
		suggestions = append(suggestions, k)
	}
	return &Completion{
		Suggestions: suggestions,
	}
}

func (a *Aliaser) Command() Command {
	aliasCompletor := &Completor{
		SuggestionFetcher: &aliasFetcher{a: a},
		Distinct:          true,
	}

	return &CommandBranch{
		Subcommands: map[string]Command{
			"a": &TerminusCommand{
				Executor: a.AddAlias,
				Args: []Arg{
					StringArg(AliasArg, true, nil),
					a.Arg,
				},
			},
			"d": &TerminusCommand{
				Executor: a.DeleteAliases,
				Args: []Arg{
					StringListArg(AliasArg, 1, UnboundedList, aliasCompletor),
				},
			},
			"g": &TerminusCommand{
				Executor: a.GetAlias,
				Args: []Arg{
					StringArg(AliasArg, true, aliasCompletor),
				},
			},
			"l": &TerminusCommand{
				Executor: a.ListAliases,
			},
			"s": &TerminusCommand{
				Executor: a.SearchAliases,
				Args: []Arg{
					StringArg(RegexpArg, true, nil),
				},
			},
		},
	}
}

// GetAlias fetches an existing alias, if it exists.
func (a *Aliaser) GetAlias(cos CommandOS, args, flags map[string]*Value, _ *OptionInfo) (*ExecutorResponse, bool) {
	alias := args[AliasArg].GetString_()
	f, ok := a.Aliases[alias]
	if !ok {
		cos.Stderr("Alias %q does not exist", alias)
		return nil, false
	}
	cos.Stdout("%s: %s", alias, f.Str())
	return nil, true
}

// AddAlias adds an alias.
func (a *Aliaser) AddAlias(cos CommandOS, args, flags map[string]*Value, _ *OptionInfo) (*ExecutorResponse, bool) {
	alias := args[AliasArg].GetString_()
	value := args[a.Arg.Name()]

	if f, ok := a.Aliases[alias]; ok {
		cos.Stderr("alias already defined: (%s: %s)", alias, f.Str())
		return nil, false
	}

	// Verify the alias.
	if a.Verifier != nil {
		if !a.Verifier(cos, alias, value, args, flags) {
			return nil, false
		}
	}

	if a.Transformer != nil {
		var ok bool
		if value, ok = a.Transformer(cos, alias, value, args, flags); !ok {
			return nil, false
		}
	}

	/*if _, err := osStat(filename); err != nil {
		cos.Stderr("file does not exist: %v", err)
		return nil, false
	}

	absPath, err := filepathAbs(filename)
	if err != nil {
		cos.Stderr("failed to get absolute file path for file %q: %v", filename, err)
		return nil, false
	}*/

	if a.Aliases == nil {
		a.Aliases = map[string]*Value{}
	}

	a.Aliases[alias] = value
	a.changed = true
	return nil, true
}

// DeleteAliases deletes an existing alias.
func (a *Aliaser) DeleteAliases(cos CommandOS, args, flags map[string]*Value, _ *OptionInfo) (*ExecutorResponse, bool) {
	for _, alias := range args[AliasArg].GetStringList().GetList() {
		if _, ok := a.Aliases[alias]; !ok {
			cos.Stderr("alias %q does not exist", alias)
		} else {
			delete(a.Aliases, alias)
			a.changed = true
		}
	}
	return nil, true
}

// ListAliases removes an existing alias.
func (a *Aliaser) ListAliases(cos CommandOS, _, _ map[string]*Value, _ *OptionInfo) (*ExecutorResponse, bool) {
	for _, aliasStr := range a.listAliases() {
		cos.Stdout(aliasStr)
	}
	return nil, true
}

func (a *Aliaser) listAliases() []string {
	keys := make([]string, 0, len(a.Aliases))
	for k := range a.Aliases {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	vs := make([]string, 0, len(keys))
	for _, k := range keys {
		vs = append(vs, fmt.Sprintf("%s: %s", k, a.Aliases[k].Str()))
	}
	return vs
}

// SearchAliases searches through existing aliases.
func (a *Aliaser) SearchAliases(cos CommandOS, args, flags map[string]*Value, _ *OptionInfo) (*ExecutorResponse, bool) {
	searchRegex, err := regexp.Compile(args[RegexpArg].GetString_())
	if err != nil {
		cos.Stderr("Invalid regexp: %v", err)
		return nil, false
	}

	for _, aliasStr := range a.listAliases() {
		if searchRegex.MatchString(aliasStr) {
			cos.Stdout(aliasStr)
		}
	}
	return nil, true
}
