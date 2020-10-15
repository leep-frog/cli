// Package commands defines all command shortcuts and sets up aliases and autocomplete functionality in bash.
package commands

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
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

// Option is a way for CLIs to define additional configuration that isn't easy
// or feasible exclusively in go.
type Option struct {
	// SetupCommand is a bash script that runs prior to the CLI. It's output
	// is passed as arguments to the command.
	SetupCommand string
}

// OptionInfo is passed to CLIs and contains info about the command's Option.
type OptionInfo struct {
	// SetupOutputFile contains the output from Option.SetupCommand
	SetupOutputFile string
}

// Command is an interface for a CLI that can be written in go.
type Command interface {
	Complete([]string) *Completion
	Execute(CommandOS, []string, *OptionInfo) (*ExecutorResponse, bool)
	Usage() []string
}

// CommandOS provides OS-related objects to executors
type CommandOS interface {
	// Writes a line to stdout.
	Stdout(string, ...interface{})
	// Writes a line to stderr.
	Stderr(string, ...interface{})
	// Close informs the os that no more data will be written.
	Close()
}

type commandOS struct {
	stdoutChan chan string
	stderrChan chan string
	wg         *sync.WaitGroup
}

func (cos *commandOS) Stdout(s string, a ...interface{}) {
	cos.stdoutChan <- fmt.Sprintf(s, a...)
}

func (cos *commandOS) Stderr(s string, a ...interface{}) {
	cos.stderrChan <- fmt.Sprintf(s, a...)
}

func (cos *commandOS) Close() {
	close(cos.stdoutChan)
	close(cos.stderrChan)
	cos.wg.Wait()
}

// NewCommandOS returns an OS that points to stdout and stderr.
func NewCommandOS() CommandOS {
	stdoutChan := make(chan string)
	stderrChan := make(chan string)
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		stdout := log.New(os.Stdout, "", 0)
		for s := range stdoutChan {
			stdout.Println(s)
		}
		wg.Done()
	}()

	go func() {
		stderr := log.New(os.Stderr, "", 0)
		for s := range stderrChan {
			stderr.Println(s)
		}
		wg.Done()
	}()
	return &commandOS{
		stdoutChan: stdoutChan,
		stderrChan: stderrChan,
		wg:         &wg,
	}
}

type TestCommandOS struct {
	stdout []string
	stderr []string
}

func (tcos *TestCommandOS) Stdout(s string, a ...interface{}) {
	if tcos.stdout == nil {
		tcos.stdout = []string{}
	}
	tcos.stdout = append(tcos.stdout, fmt.Sprintf(s, a...))
}

func (tcos *TestCommandOS) Stderr(s string, a ...interface{}) {
	if tcos.stderr == nil {
		tcos.stderr = []string{}
	}
	tcos.stderr = append(tcos.stderr, fmt.Sprintf(s, a...))
}

func (tcos *TestCommandOS) Close() {}

func (tcos *TestCommandOS) GetStdout() []string {
	return tcos.stdout
}

func (tcos *TestCommandOS) GetStderr() []string {
	return tcos.stderr
}

// ExecutorResponse is the response returned by a command.
type ExecutorResponse struct {
	// Executable is another command that should be run.
	Executable []string
}

// Executor executes a commands with the given positional arguments and flags.
type Executor func(cos CommandOS, args map[string]*Value, flags map[string]*Value, oi *OptionInfo) (*ExecutorResponse, bool)

// NoopExecutor is an Executor that does nothing.
func NoopExecutor(_ CommandOS, _ map[string]*Value, _ map[string]*Value, _ *OptionInfo) (*ExecutorResponse, bool) {
	return nil, true
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
func (cb *CommandBranch) Execute(cos CommandOS, args []string, oi *OptionInfo) (*ExecutorResponse, bool) {
	if len(args) == 0 {
		if cb.TerminusCommand == nil {
			cos.Stderr("more args required")
			return nil, false
		}
		return cb.TerminusCommand.Execute(cos, args, oi)
	}

	if sc, ok := cb.Subcommands[args[0]]; ok {
		return sc.Execute(cos, args[1:], oi)
	}

	if cb.TerminusCommand == nil {
		cos.Stderr("unknown subcommand and no terminus command defined")
		return nil, false
	}
	return cb.TerminusCommand.Execute(cos, args, oi)
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
func Execute(cos CommandOS, c Command, args []string, oi *OptionInfo) (*ExecutorResponse, bool) {
	// We don't need to parse args here because we're not doing
	// our own modification and interpretation of args like we do
	// with autocomplete.
	// TODO: check for help flag and print usage.
	return c.Execute(cos, args, oi)
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

	if completion.DontComplete {
		predictions = append(predictions, " ")
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
func (tc *TerminusCommand) Execute(cos CommandOS, args []string, oi *OptionInfo) (*ExecutorResponse, bool) {
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
			cos.Stderr("failed to process flags: %v", err)
			return nil, false
		}

		if fullyProcessed {
			flagValues[flag.Name()] = value
			idx += 1 + value.Length()
		} else {
			cos.Stderr("not enough values passed to flag %q", flag.Name())
			return nil, false
		}
	}

	// Populate args
	argIdx := 0
	populatedArgs := map[string]*Value{}
	for idx := 0; idx < len(flaglessArgs); {
		if argIdx >= len(tc.Args) {
			cos.Stderr("extra unknown args (%v)", flaglessArgs[idx:])
			return nil, false
		}

		arg := tc.Args[argIdx]
		// Ignore string values. That's only for complete.
		value, fullyProcessed, err := arg.ProcessArgs(flaglessArgs[idx:])
		if err != nil {
			cos.Stderr("failed to process args: %v", err)
			return nil, false
		}
		populatedArgs[arg.Name()] = value

		if fullyProcessed {
			idx += value.Length()
			argIdx++
		} else {
			cos.Stderr("not enough arguments for %q arg", arg.Name())
			return nil, false
		}
	}

	// Iterate to first non-optional argument
	for ; argIdx < len(tc.Args) && tc.Args[argIdx].Optional(); argIdx++ {
	}

	if argIdx != len(tc.Args) {
		nextArg := tc.Args[argIdx]
		cos.Stderr("no argument provided for %q", nextArg.Name())
		return nil, false
	}

	if tc.Executor == nil {
		cos.Stderr("no executor defined for command")
		return nil, false
	}

	return tc.Executor(cos, populatedArgs, flagValues, oi)
}

// Complete returns all possible autocomplete suggestions for the given list of arguments.
// TODO: this should return an error so it's easier to debug and test
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
