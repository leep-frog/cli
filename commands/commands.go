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

var (
	// Default is string list
	quotationChars = map[rune]bool{
		'"':  true,
		'\'': true,
	}
)

func parseArgs(unparsedArgs []string) ([]string, *rune) {
	if len(unparsedArgs) == 0 {
		return nil, nil
	}

	// Ignore if the last charater is just a quote
	var delimiterOverride *rune
	lastArg := unparsedArgs[len(unparsedArgs)-1]
	if len(lastArg) == 1 && quotationChars[rune(lastArg[0])] {
		r := rune(lastArg[0])
		delimiterOverride = &r
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

	var currentQuote *rune
	// Note: "one"two is equivalent to (onetwo) as opposed to (one two).
	for i, arg := range unparsedArgs {
		for j := 0; j < len(arg); j++ {
			char := rune(arg[j])

			if currentQuote != nil {
				if char == *currentQuote {
					currentQuote = nil
				} else {
					currentString = append(currentString, char)
				}
			} else if quotationChars[char] {
				currentQuote = &char
			} else if char == '\\' && j < len(arg)-1 && rune(arg[j+1]) == ' ' {
				currentString = append(currentString, ' ')
				j++
			} else {
				currentString = append(currentString, char)
			}
		}

		if currentQuote != nil && i != len(unparsedArgs)-1 {
			currentString = append(currentString, ' ')
		} else if len(arg) > 0 && rune(arg[len(arg)-1]) == '\\' {
			// If last character of argument is a backslash, then it's just a space
			currentString[len(currentString)-1] = ' '
		} else {
			parsedArgs = append(parsedArgs, string(currentString))
			currentString = currentString[0:0]
		}
	}

	var delimiter *rune
	if delimiterOverride != nil {
		delimiter = delimiterOverride
	} else if currentQuote != nil {
		delimiter = currentQuote
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

/*func NewTestCommandOS(t *testing.T) {
	// TODO: Add close to TODO.
}*/

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

	if completion.CaseInsenstiveSort {
		sort.Slice(predictions, func(i, j int) bool { return strings.ToLower(predictions[i]) < strings.ToLower(predictions[j]) })
	} else {
		sort.Strings(predictions)
	}
	for i, prediction := range predictions {
		if strings.Contains(prediction, " ") {
			if delimiter == nil {
				// TODO: default delimiter behavior should be defined by command?
				predictions[i] = strings.ReplaceAll(prediction, " ", "\\ ")
			} else {
				predictions[i] = fmt.Sprintf("%s%s%s", string(*delimiter), prediction, string(*delimiter))
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

func (tc *TerminusCommand) flagMap() map[string]Flag {
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
	flagMap := tc.flagMap()

	flagValues := map[string]*Value{}
	argValues := map[string]*Value{}

	// Populate flags.
	for idx := 0; idx < len(args); {
		arg := args[idx]
		flag, ok := flagMap[arg]
		if !ok {
			idx++
			continue
		}

		n, err := flag.ProcessExecuteArgs(args[(idx+1):], argValues, flagValues)
		if err != nil {
			cos.Stderr(err.Error())
			return nil, false
		}
		args = append(args[:idx], args[idx+n+1:]...)
	}

	// Populate args
	for _, arg := range tc.Args {
		n, err := arg.ProcessExecuteArgs(args, argValues, flagValues)
		if err != nil {
			cos.Stderr(err.Error())
			return nil, false
		}
		args = args[n:]
	}

	if len(args) != 0 {
		cos.Stderr("extra unknown args (%v)", args)
		return nil, false
	}

	if tc.Executor == nil {
		cos.Stderr("no executor defined for command")
		return nil, false
	}

	return tc.Executor(cos, argValues, flagValues, oi)
}

// Complete returns all possible autocomplete suggestions for the given list of arguments.
// TODO: this should return an error so it's easier to debug and test
func (tc *TerminusCommand) Complete(rawArgs []string) *Completion {
	flagMap := tc.flagMap()

	flagValues := map[string]*Value{}
	argValues := map[string]*Value{}
	args := rawArgs
	// Don't care if the last argument is a flag because
	// that is taken care of in the next step.
	for idx := 0; idx < len(args)-1; {
		arg := args[idx]
		flag, ok := flagMap[arg]
		if !ok {
			idx++
			continue
		}

		n := flag.ProcessCompleteArgs(args[(idx+1):], argValues, flagValues)
		if n+idx+1 >= len(args) {
			return flag.Complete(args[len(args)-1], argValues, flagValues)
		}
		args = append(args[:idx], args[(idx+n+1):]...)
	}

	// Check if last arg is incomplete flag
	if len(args) > 0 && strings.HasPrefix(args[len(args)-1], "-") {
		shortNames := make([]string, 0, len(tc.Flags))
		names := make([]string, 0, len(tc.Flags))
		for _, flag := range tc.Flags {
			names = append(names, fmt.Sprintf("--%s", flag.Name()))
			shortNames = append(shortNames, fmt.Sprintf("-%s", string(flag.ShortName())))
		}

		// Only show full names in this case.
		if args[len(args)-1] == "-" {
			return &Completion{
				Suggestions: filter(args, names),
			}
		}

		// Otherwise, just return all flags if the last arg is a prefix of any of them.
		return &Completion{
			Suggestions: filter(args, append(names, shortNames...)),
		}
	}

	for _, arg := range tc.Args {
		n := arg.ProcessCompleteArgs(args, argValues, flagValues)
		if n >= len(args) {
			return arg.Complete(args[len(args)-1], argValues, flagValues)
		}
		args = args[n:]
	}

	return nil
}

// Arg is a positional argument used by a TerminusCommand.
type Arg interface {
	Name() string
	ProcessCompleteArgs(rawArgs []string, args, flags map[string]*Value) int
	ProcessExecuteArgs(rawArgs []string, args, flags map[string]*Value) (int, error)
	Complete(rawValue string, args, flags map[string]*Value) *Completion
	Usage() []string
	// TODO: I believe this can be removed.
	Optional() bool
}

// Flag is a flag arguments used by a TerminusCommand.
type Flag interface {
	Name() string
	ShortName() rune
	ProcessCompleteArgs(rawArgs []string, args, flags map[string]*Value) int
	ProcessExecuteArgs(rawArgs []string, args, flags map[string]*Value) (int, error)
	Complete(rawValue string, args, flags map[string]*Value) *Completion
	Usage() []string
}
