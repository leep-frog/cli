// Package prompt allows programs to prompt the user for input.
package prompt

import (
	"bufio"
	"fmt"
	"os"
)

func PromptOptions(text string, options []string) (string, error) {
	optionMap := map[string]bool{}
	for _, o := range options {
		optionMap[o] = true
	}

	res, err := Prompt(text)
	_, ok := optionMap[res]
	for !ok && err == nil {
		fmt.Printf("Input must be one of %v, Got %q\n", options, res)
		res, err = Prompt(text)
		_, ok = optionMap[res]
	}

	if err != nil {
		res = ""
	}
	return res, err
}

func Prompt(text string) (string, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(text + ": ")
	res, err := reader.ReadString('\n')
	res = res[:(len(res) - 1)]
	if err != nil {
		return "", fmt.Errorf("failed to read from stdIn: %v", err)
	}
	return res, nil
}
