package cli

import (
	"fmt"

	"github.com/leep-frog/cli/cache"
	"github.com/leep-frog/cli/commands"
)

var (
	AllCommands = []CLI{
		//&todo.List{},
		//&emacs.Emacs{},
	}
)

type CLI interface {
	Name() string
	Alias() string
	Load(string) error
	Command() commands.Command
	Changed() bool
}

func Save(c CLI) error {
	ck := cacheKey(c)
	cash := &cache.Cache{}
	if err := cash.PutStruct(ck, c); err != nil {
		return fmt.Errorf("failed to save cli %q: %v", c.Name(), err)
	}
	return nil
}

func cacheKey(c CLI) string {
	return fmt.Sprintf("cache-key-%s", c.Name())
}

func Load(c CLI) error {
	ck := cacheKey(c)
	cash := &cache.Cache{}
	s, err := cash.Get(ck)
	if err != nil {
		return fmt.Errorf("failed to load cli %q: %v", c.Name(), err)
	}

	return c.Load(s)
}

func Execute(cli CLI, args []string) (*commands.ExecutorResponse, error) {
	if err := Load(cli); err != nil {
		return nil, fmt.Errorf("failed to load cli: %v", err)
	}
	resp, err := commands.Execute(cli.Command(), args)
	if err != nil {
		return resp, err
	}
	if cli.Changed() {
		if err := Save(cli); err != nil {
			return resp, fmt.Errorf("failed to save: %v", err)
		}
	}
	return resp, err
}

func Autocomplete(cli CLI, args []string, cursorIdx int) []string {
	if err := Load(cli); err != nil {
		return []string{fmt.Sprintf("failed to load cli: %v", err)}
	}
	return commands.Autocomplete(cli.Command(), args, cursorIdx)
}
