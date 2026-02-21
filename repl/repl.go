// Package repl implements an interactive read-eval-print loop for otto.
package repl

import (
	"fmt"
	"io"
	"os"

	"github.com/chzyer/readline"
	"github.com/robertkrimen/otto"
)

// Run starts an interactive REPL, reading JS expressions from the
// terminal and printing their results. Exit with .exit or Ctrl-D.
func Run(vm *otto.Otto) {
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "> ",
		HistoryFile:     os.TempDir() + "/gode_history",
		InterruptPrompt: "^C",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing REPL: %v\n", err)
		os.Exit(1)
	}
	defer rl.Close()

	fmt.Println("Welcome to gode – a JavaScript runtime powered by otto")
	fmt.Println("Type .exit or press Ctrl-D to quit")
	fmt.Println()

	for {
		line, err := rl.Readline()
		if err != nil {
			if err == io.EOF || err == readline.ErrInterrupt {
				fmt.Println()
				break
			}
			fmt.Fprintf(os.Stderr, "Read error: %v\n", err)
			break
		}

		if line == ".exit" {
			break
		}

		if line == "" {
			continue
		}

		value, err := vm.Run(line)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			continue
		}

		// Don't echo undefined results, same as Node.
		if !value.IsUndefined() {
			fmt.Println(value)
		}
	}
}
