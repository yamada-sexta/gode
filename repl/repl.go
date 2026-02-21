// Package repl implements an interactive read-eval-print loop for goja,
// modelled after the Node.js REPL.
package repl

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/dop251/goja"
	"github.com/dop251/goja/parser"
)

const (
	promptNormal   = "> "
	promptContinue = "... "
)

// Run starts an interactive REPL.
func Run(vm *goja.Runtime, version string) {
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

	var history []string
	var buffer string

	for {
		line, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				if buffer != "" {
					buffer = ""
					rl.SetPrompt(promptNormal)
					fmt.Println()
					continue
				}
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

		// --- Dot-commands ---
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

		if buffer != "" {
			buffer += "\n" + line
		} else {
			buffer = line
		}

		if isIncomplete(buffer) {
			rl.SetPrompt(promptContinue)
			continue
		}

		src := buffer
		history = append(history, buffer)
		buffer = ""
		rl.SetPrompt(promptNormal)

		value, err := vm.RunString(src)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			continue
		}

		if !goja.IsUndefined(value) {
			fmt.Println(value)
		}
	}
}

func isIncomplete(src string) bool {
	_, err := parser.ParseFile(nil, "", src, 0)
	if err == nil {
		return false
	}
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

func runEditorMode(rl *readline.Instance, vm *goja.Runtime, history *[]string) {
	fmt.Println("// Entering editor mode (Ctrl+D to finish, Ctrl+C to cancel)")
	var lines []string
	rl.SetPrompt("")
	for {
		line, err := rl.Readline()
		if err != nil {
			if err == io.EOF {
				break
			}
			if err == readline.ErrInterrupt {
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
	value, err := vm.RunString(src)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	if !goja.IsUndefined(value) {
		fmt.Println(value)
	}
}

func loadFile(vm *goja.Runtime, path string, history *[]string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading file: %v\n", err)
		return
	}
	src := string(data)
	*history = append(*history, src)
	fmt.Printf("// Loading %s\n", path)
	value, err := vm.RunString(src)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	if !goja.IsUndefined(value) {
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
