package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
)

const (
	AliasArg  = "ALIAS"
	RegexpArg = "REGEXP"
	FileArg   = "FILE"
)

type Aliaser interface {
	// Validate verifies the given value.
	Validate(cos CommandOS, alias string, value *Value, args, flags map[string]*Value) bool
	// Transform transforms the validated value.
	Transform(cos CommandOS, alias string, value *Value, args, flags map[string]*Value) (*Value, bool)
	// Arg is the Arg type for the alias.
	Arg() Arg
}

type AliasCLI interface {
	GetAlias(string) (*Value, bool)
	SetAlias(string, *Value)
	DeleteAlias(string)
	AllAliases() []string
}

type aliasCommand struct {
	aliaser  Aliaser
	aliasCLI AliasCLI
}

type fileAliaser struct {
	osStat  func(s string) (os.FileInfo, error)
	absPath func(s string) (string, error)
}

func NewFileAliaser() Aliaser {
	return &fileAliaser{
		osStat:  os.Stat,
		absPath: filepath.Abs,
	}
}

func TestFileAliaser(fakeStat func(s string) (os.FileInfo, error), fakeAbs func(s string) (string, error)) Aliaser {
	return &fileAliaser{
		osStat:  fakeStat,
		absPath: fakeAbs,
	}
}

func (fa *fileAliaser) Validate(cos CommandOS, alias string, value *Value, args, flags map[string]*Value) bool {
	if _, err := fa.osStat(value.GetString_()); err != nil {
		cos.Stderr("file does not exist: %v", err)
		return false
	}
	return true
}

func (fa *fileAliaser) Transform(cos CommandOS, alias string, value *Value, args, flags map[string]*Value) (*Value, bool) {
	absPath, err := fa.absPath(value.GetString_())
	if err != nil {
		cos.Stderr("failed to get absolute file path for file %q: %v", value.GetString_(), err)
		return nil, false
	}

	return StringValue(absPath), true
}

func (*fileAliaser) Arg() Arg {
	completor := &Completor{
		SuggestionFetcher: &FileFetcher{},
	}
	return StringArg(FileArg, true, completor)
}

type AliasFetcher struct {
	ac *aliasCommand
}

func (af *AliasFetcher) Fetch(value *Value, args, flags map[string]*Value) *Completion {
	suggestions := make([]string, 0, len(af.ac.aliasCLI.AllAliases()))
	for _, k := range af.ac.aliasCLI.AllAliases() {
		suggestions = append(suggestions, k)
	}
	return &Completion{
		Suggestions: suggestions,
	}
}

func AliasSubcommands(cli AliasCLI, aliaser Aliaser) map[string]Command {
	ac := &aliasCommand{
		aliasCLI: cli,
		aliaser:  aliaser,
	}
	aliasCompletor := &Completor{
		SuggestionFetcher: &AliasFetcher{ac: ac},
		Distinct:          true,
	}

	return map[string]Command{
		"a": &TerminusCommand{
			Executor: ac.AddAlias,
			Args: []Arg{
				StringArg(AliasArg, true, nil),
				ac.aliaser.Arg(),
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
	}
}

// GetAlias fetches an existing alias, if it exists.
func (ac *aliasCommand) GetAlias(cos CommandOS, args, flags map[string]*Value, _ *OptionInfo) (*ExecutorResponse, bool) {
	alias := args[AliasArg].GetString_()
	f, ok := ac.aliasCLI.GetAlias(alias)
	if !ok {
		cos.Stderr("Alias %q does not exist", alias)
		return nil, false
	}
	cos.Stdout("%s: %s", alias, f.Str())
	return nil, true
}

// AddAlias adds an alias.
func (ac *aliasCommand) AddAlias(cos CommandOS, args, flags map[string]*Value, _ *OptionInfo) (*ExecutorResponse, bool) {
	alias := args[AliasArg].GetString_()
	value := args[ac.aliaser.Arg().Name()]

	if f, ok := ac.aliasCLI.GetAlias(alias); ok {
		cos.Stderr("alias already defined: (%s: %s)", alias, f.Str())
		return nil, false
	}

	// Verify the alias.
	if !ac.aliaser.Validate(cos, alias, value, args, flags) {
		return nil, false
	}

	var ok bool
	if value, ok = ac.aliaser.Transform(cos, alias, value, args, flags); !ok {
		return nil, false
	}

	ac.aliasCLI.SetAlias(alias, value)
	return nil, true
}

// DeleteAliases deletes an existing alias.
func (ac *aliasCommand) DeleteAliases(cos CommandOS, args, flags map[string]*Value, _ *OptionInfo) (*ExecutorResponse, bool) {
	for _, alias := range args[AliasArg].GetStringList().GetList() {
		if _, ok := ac.aliasCLI.GetAlias(alias); !ok {
			cos.Stderr("alias %q does not exist", alias)
		} else {
			ac.aliasCLI.DeleteAlias(alias)
		}
	}
	return nil, true
}

// ListAliases removes an existing alias.
func (ac *aliasCommand) ListAliases(cos CommandOS, _, _ map[string]*Value, _ *OptionInfo) (*ExecutorResponse, bool) {
	for _, aliasStr := range ac.listAliases() {
		cos.Stdout(aliasStr)
	}
	return nil, true
}

func (ac *aliasCommand) listAliases() []string {
	keys := make([]string, 0, len(ac.aliasCLI.AllAliases()))
	for _, k := range ac.aliasCLI.AllAliases() {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	vs := make([]string, 0, len(keys))
	for _, k := range keys {
		v, _ := ac.aliasCLI.GetAlias(k)
		vs = append(vs, fmt.Sprintf("%s: %s", k, v.Str()))
	}
	return vs
}

// SearchAliases searches through existing aliases.
func (ac *aliasCommand) SearchAliases(cos CommandOS, args, flags map[string]*Value, _ *OptionInfo) (*ExecutorResponse, bool) {
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
