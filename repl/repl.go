// Package repl implements an interactive read-eval-print loop for otto,
// modelled after the Node.js REPL. It supports dot-commands (.help,
// .break, .clear, .editor, .load, .save, .exit), multi-line input for
// incomplete expressions, and Ctrl-C / Ctrl-D handling.
package repl

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/robertkrimen/otto"
	"github.com/robertkrimen/otto/parser"

	"gode/compat"
)

const (
	promptNormal   = "> "
	promptContinue = "... "
)

// Run starts an interactive REPL, reading JS expressions from the
// terminal and printing their results.
func Run(vm *otto.Otto, version string) {
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          promptNormal,
		HistoryFile:     os.TempDir() + "/gode_history",
		InterruptPrompt: "",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing REPL: %v\n", err)
		os.Exit(1)
	}
	defer rl.Close()

	fmt.Printf("Welcome to gode %s.\n", version)
	fmt.Println(`Type ".help" for more information.`)

	var history []string // all successfully evaluated lines for .save
	var buffer string    // accumulator for multi-line input

	for {
		line, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				// Ctrl-C: if we have a partial expression, discard it.
				if buffer != "" {
					buffer = ""
					rl.SetPrompt(promptNormal)
					fmt.Println()
					continue
				}
				// Otherwise print hint.
				fmt.Println()
				fmt.Println(`(To exit, press Ctrl+D or type .exit)`)
				continue
			}
			if err == io.EOF {
				fmt.Println()
				break
			}
			fmt.Fprintf(os.Stderr, "Read error: %v\n", err)
			break
		}

		// --- Dot-commands (only when not in multi-line mode) ---
		if buffer == "" && len(line) > 0 && line[0] == '.' {
			cmd, arg := parseDotCommand(line)
			switch cmd {
			case ".exit":
				return
			case ".help":
				printHelp()
				continue
			case ".break", ".clear":
				buffer = ""
				rl.SetPrompt(promptNormal)
				continue
			case ".editor":
				runEditorMode(rl, vm, &history)
				continue
			case ".load":
				if arg == "" {
					fmt.Fprintln(os.Stderr, "Usage: .load <filename>")
				} else {
					loadFile(vm, arg, &history)
				}
				continue
			case ".save":
				if arg == "" {
					fmt.Fprintln(os.Stderr, "Usage: .save <filename>")
				} else {
					saveSession(arg, history)
				}
				continue
			default:
				fmt.Fprintf(os.Stderr, "Invalid REPL keyword: %s\n", cmd)
				continue
			}
		}

		if line == "" && buffer == "" {
			continue
		}

		// Accumulate input.
		if buffer != "" {
			buffer += "\n" + line
		} else {
			buffer = line
		}

		// Check if the expression looks incomplete.
		if isIncomplete(buffer) {
			rl.SetPrompt(promptContinue)
			continue
		}

		// Evaluate the complete expression.
		src := compat.Transform(buffer)
		history = append(history, buffer)
		buffer = ""
		rl.SetPrompt(promptNormal)

		value, err := vm.Run(src)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			continue
		}

		if !value.IsUndefined() {
			fmt.Println(value)
		}
	}
}

// isIncomplete returns true if src appears to be a partial expression
// that needs more input (unclosed braces, parens, brackets, or strings).
func isIncomplete(src string) bool {
	_, err := parser.ParseFile(nil, "", compat.Transform(src), 0)
	if err == nil {
		return false
	}
	// otto's parser returns "Unexpected end of input" for incomplete code.
	msg := err.Error()
	return strings.Contains(msg, "Unexpected end of input") ||
		strings.Contains(msg, "unexpected end")
}

func parseDotCommand(line string) (cmd, arg string) {
	parts := strings.SplitN(strings.TrimSpace(line), " ", 2)
	cmd = parts[0]
	if len(parts) > 1 {
		arg = strings.TrimSpace(parts[1])
	}
	return
}

func printHelp() {
	fmt.Println(`.break    Sometimes you get stuck, this gets you out`)
	fmt.Println(`.clear    Alias for .break`)
	fmt.Println(`.editor   Enter editor mode`)
	fmt.Println(`.exit     Exit the REPL`)
	fmt.Println(`.help     Print this help message`)
	fmt.Println(`.load     Load JS from a file into the REPL session`)
	fmt.Println(`.save     Save all evaluated commands in this REPL session to a file`)
	fmt.Println()
	fmt.Println(`Press Ctrl+C to abort current expression, Ctrl+D to exit the REPL`)
}

func runEditorMode(rl *readline.Instance, vm *otto.Otto, history *[]string) {
	fmt.Println("// Entering editor mode (Ctrl+D to finish, Ctrl+C to cancel)")
	var lines []string
	rl.SetPrompt("")
	for {
		line, err := rl.Readline()
		if err != nil {
			if err == io.EOF {
				// Ctrl+D: evaluate accumulated code.
				break
			}
			if err == readline.ErrInterrupt {
				// Ctrl+C: cancel editor mode.
				fmt.Println()
				rl.SetPrompt(promptNormal)
				return
			}
			break
		}
		lines = append(lines, line)
	}
	rl.SetPrompt(promptNormal)
	if len(lines) == 0 {
		return
	}

	src := strings.Join(lines, "\n")
	*history = append(*history, src)
	value, err := vm.Run(compat.Transform(src))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	if !value.IsUndefined() {
		fmt.Println(value)
	}
}

func loadFile(vm *otto.Otto, path string, history *[]string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading file: %v\n", err)
		return
	}
	src := string(data)
	*history = append(*history, src)
	fmt.Printf("// Loading %s\n", path)
	value, err := vm.Run(compat.Transform(src))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	if !value.IsUndefined() {
		fmt.Println(value)
	}
}

func saveSession(path string, history []string) {
	content := strings.Join(history, "\n") + "\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving session: %v\n", err)
		return
	}
	fmt.Printf("Session saved to: %s\n", path)
}
