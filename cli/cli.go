package cli

import (
	"fmt"

	"github.com/leep-frog/cli/cache"
	"github.com/leep-frog/cli/commands"
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

func Execute(cos commands.CommandOS, cli CLI, args []string) (*commands.ExecutorResponse, bool) {
	if err := Load(cli); err != nil {
		cos.Stderr("failed to load cli: %v", err)
		return nil, false
	}
	resp, ok := commands.Execute(cos, cli.Command(), args)
	if !ok {
		return resp, ok
	}
	if cli.Changed() {
		if err := Save(cli); err != nil {
			cos.Stderr("failed to save CLI data: %v", err)
			return resp, false
		}
	}
	return resp, true
}

func Autocomplete(cli CLI, args []string, cursorIdx int) []string {
	if err := Load(cli); err != nil {
		return []string{fmt.Sprintf("failed to load cli: %v", err)}
	}
	return commands.Autocomplete(cli.Command(), args, cursorIdx)
}
