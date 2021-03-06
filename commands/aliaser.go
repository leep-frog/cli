package commands

import (
	"fmt"
	"os"
	"regexp"
	"sort"
)

const (
	AliasArg  = "ALIAS"
	RegexpArg = "REGEXP"
)

var (
	osStat = os.Stat
)

type Aliaser interface {
	// Validate verifies the given value.
	Validate(cos CommandOS, alias string, value *Value, args, flags map[string]*Value) bool
	// Transform transforms the validated value.
	Transform(cos CommandOS, alias string, value *Value, args, flags map[string]*Value) (*Value, bool)
	// Arg is the Arg type for the alias.
	Arg() Arg
}

type AliasCommand struct {
	Aliaser Aliaser

	Aliases map[string]*Value

	changed bool
}

/*func FileVerifier(cos CommandOS, alias string, value *Value, args, flags map[string]*Value) bool {
	if !value.IsType(StringType) {
		cos.Stderr("file verifier requires a string type")
		return false
	}

	if _, err := osStat(value.GetString_()); err != nil {
		cos.Stderr("file does not exist: %v", err)
		return false
	}

	return true
}

func FileTransformer(cos CommandOS, alias string, value *Value, args, flags map[string]*Value) (*Value, bool) {
	absPath, err := filepathAbs(value.GetString_())
	if err != nil {
		cos.Stderr("failed to get absolute file path for file %q: %v", filename, err)
		return false
	}

	return stringVal(absPath)
	return nil, false
}*/

type AliasFetcher struct {
	ac *AliasCommand
}

func (ac *AliasCommand) Changed() bool {
	return ac.changed
}

func (af *AliasFetcher) Fetch(value *Value, args, flags map[string]*Value) *Completion {
	suggestions := make([]string, 0, len(af.ac.Aliases))
	for k := range af.ac.Aliases {
		suggestions = append(suggestions, k)
	}
	return &Completion{
		Suggestions: suggestions,
	}
}

func (ac *AliasCommand) Command() Command {
	aliasCompletor := &Completor{
		SuggestionFetcher: &AliasFetcher{ac: ac},
		Distinct:          true,
	}

	return &CommandBranch{
		Subcommands: map[string]Command{
			"a": &TerminusCommand{
				Executor: ac.AddAlias,
				Args: []Arg{
					StringArg(AliasArg, true, nil),
					ac.Aliaser.Arg(),
				},
			},
			"d": &TerminusCommand{
				Executor: ac.DeleteAliases,
				Args: []Arg{
					StringListArg(AliasArg, 1, UnboundedList, aliasCompletor),
				},
			},
			"g": &TerminusCommand{
				Executor: ac.GetAlias,
				Args: []Arg{
					StringArg(AliasArg, true, aliasCompletor),
				},
			},
			"l": &TerminusCommand{
				Executor: ac.ListAliases,
			},
			"s": &TerminusCommand{
				Executor: ac.SearchAliases,
				Args: []Arg{
					StringArg(RegexpArg, true, nil),
				},
			},
		},
	}
}

// GetAlias fetches an existing alias, if it exists.
func (ac *AliasCommand) GetAlias(cos CommandOS, args, flags map[string]*Value, _ *OptionInfo) (*ExecutorResponse, bool) {
	alias := args[AliasArg].GetString_()
	f, ok := ac.Aliases[alias]
	if !ok {
		cos.Stderr("Alias %q does not exist", alias)
		return nil, false
	}
	cos.Stdout("%s: %s", alias, f.Str())
	return nil, true
}

// AddAlias adds an alias.
func (ac *AliasCommand) AddAlias(cos CommandOS, args, flags map[string]*Value, _ *OptionInfo) (*ExecutorResponse, bool) {
	alias := args[AliasArg].GetString_()
	value := args[ac.Aliaser.Arg().Name()]

	if f, ok := ac.Aliases[alias]; ok {
		cos.Stderr("alias already defined: (%s: %s)", alias, f.Str())
		return nil, false
	}

	// Verify the alias.
	if !ac.Aliaser.Validate(cos, alias, value, args, flags) {
		return nil, false
	}

	var ok bool
	if value, ok = ac.Aliaser.Transform(cos, alias, value, args, flags); !ok {
		return nil, false
	}

	if ac.Aliases == nil {
		ac.Aliases = map[string]*Value{}
	}

	ac.Aliases[alias] = value
	ac.changed = true
	return nil, true
}

// DeleteAliases deletes an existing alias.
func (ac *AliasCommand) DeleteAliases(cos CommandOS, args, flags map[string]*Value, _ *OptionInfo) (*ExecutorResponse, bool) {
	for _, alias := range args[AliasArg].GetStringList().GetList() {
		if _, ok := ac.Aliases[alias]; !ok {
			cos.Stderr("alias %q does not exist", alias)
		} else {
			delete(ac.Aliases, alias)
			ac.changed = true
		}
	}
	return nil, true
}

// ListAliases removes an existing alias.
func (ac *AliasCommand) ListAliases(cos CommandOS, _, _ map[string]*Value, _ *OptionInfo) (*ExecutorResponse, bool) {
	for _, aliasStr := range ac.listAliases() {
		cos.Stdout(aliasStr)
	}
	return nil, true
}

func (ac *AliasCommand) listAliases() []string {
	keys := make([]string, 0, len(ac.Aliases))
	for k := range ac.Aliases {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	vs := make([]string, 0, len(keys))
	for _, k := range keys {
		vs = append(vs, fmt.Sprintf("%s: %s", k, ac.Aliases[k].Str()))
	}
	return vs
}

// SearchAliases searches through existing aliases.
func (ac *AliasCommand) SearchAliases(cos CommandOS, args, flags map[string]*Value, _ *OptionInfo) (*ExecutorResponse, bool) {
	searchRegex, err := regexp.Compile(args[RegexpArg].GetString_())
	if err != nil {
		cos.Stderr("Invalid regexp: %v", err)
		return nil, false
	}

	for _, aliasStr := range ac.listAliases() {
		if searchRegex.MatchString(aliasStr) {
			cos.Stdout(aliasStr)
		}
	}
	return nil, true
}
