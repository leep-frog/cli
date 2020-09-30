package cli

import (
	"fmt"
	"io/ioutil"
)

const (
	autocompleteFunction = `
	function _custom_autocomplete {
	  tFile="$(mktemp)"
	  tFile2="$(mktemp)"
	  echo go run github.com/leep-frog/cli/cmd/autocomplete $COMP_CWORD $COMP_LINE | sed 's/"/\\"/g' | sed "s/'/\\\\'/g" > $tFile
	  chmod +x $tFile
	  bash $tFile | sed 's/"/\"/g' | sed "s/'/\\\\'/g" > $tFile2
	  local IFS=$'\n'
	  COMPREPLY=( $(compgen -W "$(cat $tFile2)" ""))
		rm $tFile $tFile2
	}

	`
)

func (cli *CLI) Source() {
	f, err := ioutil.TempFile("", "golang-cli-source")
	if err != nil {
		// STDOUT: fmt.Println("Failed to create tmp file: %v", err)
		fmt.Println("badness", err)
		return
	}

	f.WriteString(autocompleteFunction)
	for _, cmd := range cli.AllCommands {
		alias := cmd.Alias()
		if _, err := f.WriteString(fmt.Sprintf("alias %s=\"go run github.com/leep-frog/cli/cmd/run %s\"\n", alias, alias)); err != nil {
			// STDOUT: fmt.Println("failed to write alias to file: %v", err)
			fmt.Println("badness 1", err)
			return
		}
		if _, err := f.WriteString(fmt.Sprintf("complete -F _custom_autocomplete %s\n", alias)); err != nil {
			// fmt.Println("failed to write autocomplete: %v", err): STDOUT
			fmt.Println("badness 2", err)
			return
		}
	}
	fmt.Println(f.Name())
	// TODO: find better way to get paths instead of relying on environment variables
	//f.WriteString(fmt.Sprintf("alias %s=\"go run github.com/leep-frog/cli/cmd/run %s\"\n", alias, alias))
	// fmt.Printf("alias gt=\"pushd . ; cd $OPT_GO ; go test ./... ; popd\"")
}
