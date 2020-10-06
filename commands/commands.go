// Package commands defines all command shortcuts and sets up aliases and autocomplete functionality in bash.
package commands

import (
	"fmt"
	"sort"
	"strings"
)

func parseArgs(unparsedArgs []string) ([]string, *string) {
	// Ignore if the last charater is just a quote
	var delimiterOverride string
	if len(unparsedArgs) > 0 && (unparsedArgs[len(unparsedArgs)-1] == "\"" || unparsedArgs[len(unparsedArgs)-1] == "'") {
		delimiterOverride = unparsedArgs[len(unparsedArgs)-1]
		unparsedArgs[len(unparsedArgs)-1] = ""
	}

	// Words might be combined so parsed args will be less than or equal to unparsedArgs length.
	parsedArgs := make([]string, 0, len(unparsedArgs))

	// Max length of the string can be all characters (including spaces).
	totalLen := len(unparsedArgs)
	for _, arg := range unparsedArgs {
		totalLen += len(arg)
	}
	currentString := make([]rune, 0, totalLen)

	// TODO: this should be an enum (iota)
	inSingle, inDouble := false, false
	// Note: "one"two is equivalent to (onetwo) as opposed to (one two).
	for i, arg := range unparsedArgs {
		for j := 0; j < len(arg); j++ {
			char := rune(arg[j])

			if inSingle {
				if char == '\'' {
					inSingle = false
				} else {
					currentString = append(currentString, char)
				}
			} else if inDouble {
				if char == '"' {
					inDouble = false
				} else {
					currentString = append(currentString, char)
				}
			} else if char == '\'' {
				inSingle = true
			} else if char == '"' {
				inDouble = true
			} else if char == '\\' && j < len(arg)-1 && rune(arg[j+1]) == ' ' {
				currentString = append(currentString, ' ')
				j++
			} else {
				currentString = append(currentString, char)
			}
		}

		if (inSingle || inDouble) && i != len(unparsedArgs)-1 {
			currentString = append(currentString, ' ')
		} else if len(arg) > 0 && rune(arg[len(arg)-1]) == '\\' {
			// If last character of argument is a backslash, then it's just a space
			currentString[len(currentString)-1] = ' '
		} else {
			parsedArgs = append(parsedArgs, string(currentString))
			currentString = currentString[0:0]
		}
	}

	var delimiter *string
	if delimiterOverride != "" {
		delimiter = &delimiterOverride
	} else if inDouble {
		dq := `"`
		delimiter = &dq
	} else if inSingle {
		sq := "'"
		delimiter = &sq
	}

	return parsedArgs, delimiter
}

// Command is an interface for a CLI that can be written in go.
type Command interface {
	Complete([]string) *Completion
	Execute([]string) (*ExecutorResponse, error)
	Usage() []string
}

// ExecutorResponse is the response returned by a command.
type ExecutorResponse struct {
	// Stdout is the output that should be sent to stdout.
	Stdout []string
	// Stderr is the output that should be sent to stderr.
	Stderr []string
	// Executable is another command that should be run.
	Executable []string
}

// Executor executes a commands with the given positional arguments and flags.
type Executor func(args map[string]*Value, flags map[string]*Value) (*ExecutorResponse, error)

// NoopExecutor is an Executor that does nothing.
func NoopExecutor(_ map[string]*Value, _ map[string]*Value) (*ExecutorResponse, error) {
	return nil, nil
}

// TODO: combine terminusCommand and commandBranch??
// TerminusCommand is a command that processes dynamic arguments and flags.
type TerminusCommand struct {
	Args     []Arg
	Flags    []Flag
	Executor Executor
}

// CommandBranch is a command that splits into other commands depending on positional arguments.
type CommandBranch struct {
	Subcommands                  map[string]Command
	TerminusCommand              *TerminusCommand
	IgnoreSubcommandAutocomplete bool
}

// Usage returns the usage info
func (cb *CommandBranch) Usage() []string {
	usage := make([]string, 0, len(cb.Subcommands)*5)

	keys := make([]string, 0, len(cb.Subcommands))
	for k := range cb.Subcommands {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// TODO: add or (|) symbols.
	for _, k := range keys {
		v := cb.Subcommands[k]
		usage = append(usage, k)
		usage = append(usage, v.Usage()...)
		usage = append(usage, "\n")
	}

	if cb.TerminusCommand == nil {
		return usage
	}
	return append(usage, cb.TerminusCommand.Usage()...)
}

// Execute executes the corresponding subcommand.
func (cb *CommandBranch) Execute(args []string) (*ExecutorResponse, error) {
	if len(args) == 0 {
		if cb.TerminusCommand == nil {
			return nil, fmt.Errorf("more args required: USAGE TODO")
		}
		return cb.TerminusCommand.Execute(args)
	}

	if sc, ok := cb.Subcommands[args[0]]; ok {
		return sc.Execute(args[1:])
	}

	if cb.TerminusCommand == nil {
		return nil, fmt.Errorf("unknown subcommand and no terminus command defined: USAGE TODO")
	}
	return cb.TerminusCommand.Execute(args)
}

// Complete returns autocomplete suggestions.
func (cb *CommandBranch) Complete(args []string) *Completion {
	// Return subcommands and terminus command suggestions if only one argument.
	if len(args) <= 1 {
		suggestions := make([]string, 0, len(cb.Subcommands))

		if !cb.IgnoreSubcommandAutocomplete {
			for k := range cb.Subcommands {
				suggestions = append(suggestions, k)
			}
			suggestions = filter(args, suggestions)
		}

		if cb.TerminusCommand != nil {
			// the autocomplete command will filter if needed
			c := cb.TerminusCommand.Complete(args)
			if c == nil {
				c = &Completion{}
			}
			c.Suggestions = append(c.Suggestions, suggestions...)
			return c
		}

		return &Completion{
			Suggestions: suggestions,
		}
	}

	// If first argument is a subcommand, then return it's suggestions
	if sc, ok := cb.Subcommands[args[0]]; ok {
		return sc.Complete(args[1:])
	}

	// Otherwise, we only have the terminus command left.
	if cb.TerminusCommand != nil {
		return cb.TerminusCommand.Complete(args)
	}

	return nil
}

// Execute executes the given unparsed command.
// TODO: this should only return an executor response
func Execute(c Command, unparsedArgs []string) (*ExecutorResponse, error) {
	// TODO: check for help flag and print usage.
	args, _ := parseArgs(unparsedArgs)

	return c.Execute(args)
}

func filter(args, suggestions []string) []string {
	if len(args) == 0 {
		return suggestions
	}

	lastArg := args[len(args)-1]
	var filtered []string
	for _, arg := range suggestions {
		if strings.HasPrefix(arg, lastArg) {
			filtered = append(filtered, arg)
		}
	}
	return filtered
}

// Autocomplete completes the given unparsed command.
func Autocomplete(c Command, unparsedArgs []string, cursorIdx int) []string {
	args, delimiter := parseArgs(unparsedArgs)

	if cursorIdx > len(args) || len(args) == 0 {
		args = append(args, "")
	}

	completion := c.Complete(args)
	if completion == nil {
		completion = &Completion{}
	}
	predictions := completion.Suggestions

	sort.Strings(predictions)
	for i, prediction := range predictions {
		if strings.Contains(prediction, " ") {
			if delimiter == nil {
				// TODO: default delimiter behavior should be defined by command?
				predictions[i] = strings.ReplaceAll(prediction, " ", "\\ ")
			} else {
				predictions[i] = fmt.Sprintf("%s%s%s", *delimiter, prediction, *delimiter)
			}
		}
	}
	return predictions
}

// Usage returns usage info about the command.
func (tc *TerminusCommand) Usage() []string {
	usage := make([]string, 0, 3*(len(tc.Args)+len(tc.Flags)))
	for _, a := range tc.Args {
		usage = append(usage, a.Usage()...)
	}
	for _, f := range tc.Flags {
		usage = append(usage, f.Usage()...)
	}
	return usage
}

func (tc *TerminusCommand) flagMap(args []string) map[string]Flag {
	flagMap := map[string]Flag{}
	for _, flag := range tc.Flags {
		flagMap[fmt.Sprintf("--%s", flag.Name())] = flag
		if flag.ShortName() != 0 {
			flagMap[fmt.Sprintf("-%c", flag.ShortName())] = flag
		}
	}
	return flagMap
}

// Execute loads flags and args and then runs it's executor.
func (tc *TerminusCommand) Execute(args []string) (*ExecutorResponse, error) {
	flagMap := tc.flagMap(args)

	flagValues := map[string]*Value{}
	flaglessArgs := make([]string, 0, len(args))
	for idx := 0; idx < len(args); {
		arg := args[idx]
		flag, ok := flagMap[arg]
		if !ok {
			flaglessArgs = append(flaglessArgs, arg)
			idx++
			continue
		}

		// Ignore string values. That's only for complete.
		value, fullyProcessed, err := flag.ProcessArgs(args[(idx + 1):])
		if err != nil {
			return nil, fmt.Errorf("failed to process flags: %v", err)
		}

		if fullyProcessed {
			flagValues[flag.Name()] = value
			idx += 1 + value.Length()
		} else {
			return nil, fmt.Errorf("not enough values passed to flag %q: TODO USAGE", flag.Name())
		}
	}

	// Populate args
	// TODO: populate specific types?
	argIdx := 0
	populatedArgs := map[string]*Value{}
	for idx := 0; idx < len(flaglessArgs); {
		if argIdx >= len(tc.Args) {
			return nil, fmt.Errorf("extra unknown args (%v)", flaglessArgs[idx:])
		}

		arg := tc.Args[argIdx]
		// Ignore string values. That's only for complete.
		value, fullyProcessed, err := arg.ProcessArgs(flaglessArgs[idx:])
		if err != nil {
			return nil, fmt.Errorf("failed to process args: %v", err)
		}
		populatedArgs[arg.Name()] = value

		if fullyProcessed {
			idx += value.Length()
			argIdx++
		} else {
			return nil, fmt.Errorf("not enough arguments for %q arg", arg.Name())
		}
	}

	// Iterate to first non-optional argument
	for ; argIdx < len(tc.Args) && tc.Args[argIdx].Optional(); argIdx++ {
	}

	if argIdx != len(tc.Args) {
		nextArg := tc.Args[argIdx]
		return nil, fmt.Errorf("no argument provided for %q", nextArg.Name())
	}

	if tc.Executor == nil {
		return nil, fmt.Errorf("no executor defined for command")
	}

	return tc.Executor(populatedArgs, flagValues)
}

// Complete returns all possible autocomplete suggestions for the given list of arguments.
func (tc *TerminusCommand) Complete(args []string) *Completion {
	// TODO: combine common logic between this and Execute
	flagMap := tc.flagMap(args)

	// TODO: short boolean flags should be combinable (`grep -or ...` for example)

	flagValues := map[string]*Value{}
	flaglessArgs := make([]string, 0, len(args))
	for idx := 0; idx < len(args); {
		arg := args[idx]
		flag, ok := flagMap[arg]
		if !ok {
			flaglessArgs = append(flaglessArgs, arg)
			idx++
			continue
		}

		// If we're at the last arg, then just return all flags (and let filter take care of the rest)
		if idx == len(args)-1 {
			allFlags := make([]string, 0, len(flagMap))
			for k := range flagMap {
				allFlags = append(allFlags, k)
			}
			return &Completion{
				Suggestions: filter(args, allFlags),
			}
		}

		value, fullyProcessed, _ := flag.ProcessArgs(args[(idx + 1):])

		flagValues[flag.Name()] = value
		if fullyProcessed {
			idx += value.Length() + 1 // + 1 for flag itself
			if idx >= len(args) {
				return flag.Complete(nil, flagValues)
			}
		} else {
			return flag.Complete(nil, flagValues)
		}
	}

	// Check if last arg is incomplete flag
	if len(flaglessArgs) > 0 {
		lastArg := flaglessArgs[len(flaglessArgs)-1]

		if lastArg == "" {
			goto positional
		}

		// Only show full flag names if just a hyphen
		if lastArg == "-" {
			fullFlags := make([]string, 0, len(tc.Flags))
			for _, flag := range tc.Flags {
				fullFlags = append(fullFlags, fmt.Sprintf("--%s", flag.Name()))
			}
			return &Completion{
				Suggestions: filter(args, fullFlags),
			}
		}

		// Otherwise, just return all flags if the last arg is a prefix of any of them.
		matches := false
		allFlags := make([]string, 0, len(flagMap))
		for k := range flagMap {
			matches = matches || strings.HasPrefix(k, lastArg)
			allFlags = append(allFlags, k)
		}
		if matches {
			return &Completion{
				Suggestions: filter(args, allFlags),
			}
		}
	}

positional:

	if len(tc.Args) == 0 {
		return nil
	}

	argIdx := 0
	populatedArgs := map[string]*Value{}
	for idx := 0; idx < len(flaglessArgs); {
		if argIdx >= len(tc.Args) {
			return nil
		}

		arg := tc.Args[argIdx]
		value, fullyProcessed, _ := arg.ProcessArgs(flaglessArgs[idx:])
		populatedArgs[arg.Name()] = value

		if fullyProcessed {
			idx += value.Length()
			// if we are out of args then we should autocomplete the given arg.
			if idx >= len(flaglessArgs) {
				break
			}
			argIdx++
		} else {
			break
		}
	}

	// TODO: ignore the last value?
	return tc.Args[argIdx].Complete(populatedArgs, flagValues)
}

// TODO: value options

// Arg is a positional argument used by a TerminusCommand.
type Arg interface {
	Name() string
	ProcessArgs(args []string) (*Value, bool, error)
	Complete(args, flags map[string]*Value) *Completion
	Usage() []string
	Optional() bool
}

// Flag is a flag arguments used by a TerminusCommand.
type Flag interface {
	Name() string
	ShortName() rune
	ProcessArgs(args []string) (*Value, bool, error)
	Complete(args, flags map[string]*Value) *Completion
	Usage() []string
}
