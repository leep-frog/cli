package color

import (
	"github.com/leep-frog/commands/commands"
)

type fetcher struct{}

// TODO: add existing stuff in here so don't display already present format.
func (f *fetcher) Fetch(value *commands.Value, args, flags map[string]*commands.Value) *commands.Completion {
	return &commands.Completion{
		Suggestions: Attributes(),
	}
}

func Completor() *commands.Completor {
	return &commands.Completor{
		Distinct:          true,
		SuggestionFetcher: &fetcher{},
	}
}

var (
	ArgName = "format"
	Arg     = commands.StringListArg(ArgName, 1, commands.UnboundedList, Completor())
)

// TODO: have this accept commandOS and write to stderr with any issues
func ApplyCodes(f *Format, args map[string]*commands.Value) (*Format, bool) {
	if f == nil {
		f = &Format{}
	}
	codes := args[ArgName].StringList()
	for _, c := range codes {
		f.AddAttribute(c)
	}
	return f, len(codes) != 0
}
