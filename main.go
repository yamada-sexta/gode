package main

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/dop251/goja"

	"gode/envfile"
	"gode/modules"
	"gode/process"
	"gode/repl"
	"gode/runner"
)

// Options holds the parsed CLI flags and positional arguments.
type Options struct {
	Version         bool
	Help            bool
	Eval            string
	Print           string
	Check           bool
	Interactive     bool
	EnvFiles        []string
	EnvFilesIfExist []string
	ReadStdin       bool
	Script          string
	ScriptArgs      []string
}

func main() {
	opts, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "gode: %v\n", err)
		os.Exit(9) // Node uses 9 for invalid arguments
	}

	if opts.Version {
		fmt.Println(getVersion())
		return
	}

	if opts.Help {
		printHelp()
		return
	}

	// Load environment files before anything else so that the
	// variables are available to the JS runtime.
	for _, f := range opts.EnvFiles {
		if err := envfile.Load(f); err != nil {
			fmt.Fprintf(os.Stderr, "gode: %v\n", err)
			os.Exit(1)
		}
	}
	for _, f := range opts.EnvFilesIfExist {
		if err := envfile.LoadIfExists(f); err != nil {
			fmt.Fprintf(os.Stderr, "gode: %v\n", err)
			os.Exit(1)
		}
	}

	// --check: parse only, no execution.
	if opts.Check {
		if opts.Script == "" {
			fmt.Fprintf(os.Stderr, "gode: --check requires a file argument\n")
			os.Exit(1)
		}
		runner.CheckSyntax(opts.Script)
		return
	}

	// From here on we need a VM.
	vm := goja.New()
	process.Setup(vm, getVersion(), opts.Script, opts.ScriptArgs)
	modules.NewLoader(vm)

	// -e / --eval
	if opts.Eval != "" {
		runner.RunEval(vm, opts.Eval)
		if opts.Interactive {
			repl.Run(vm, getVersion())
		}
		return
	}

	// -p / --print
	if opts.Print != "" {
		runner.RunPrint(vm, opts.Print)
		if opts.Interactive {
			repl.Run(vm, getVersion())
		}
		return
	}

	// - (stdin)
	if opts.ReadStdin {
		runner.RunStdin(vm)
		return
	}

	// script.js [args...]
	if opts.Script != "" {
		runner.RunFile(vm, opts.Script)
		if opts.Interactive {
			repl.Run(vm, getVersion())
		}
		return
	}

	// Default: interactive REPL.
	repl.Run(vm, getVersion())
}

// parseArgs processes the raw CLI arguments following Node.js conventions:
// flags are consumed until a non-flag positional (the script) is found;
// everything after the script is passed through as script arguments.
func parseArgs(args []string) (*Options, error) {
	opts := &Options{}
	i := 0

	for i < len(args) {
		arg := args[i]

		// "--" ends option parsing.
		if arg == "--" {
			i++
			if opts.Script == "" && i < len(args) {
				opts.Script = args[i]
				i++
			}
			opts.ScriptArgs = append(opts.ScriptArgs, args[i:]...)
			break
		}

		// "-" means read from stdin.
		if arg == "-" {
			opts.ReadStdin = true
			i++
			opts.ScriptArgs = append(opts.ScriptArgs, args[i:]...)
			break
		}

		// First non-flag argument is the script.
		if !strings.HasPrefix(arg, "-") {
			opts.Script = arg
			i++
			opts.ScriptArgs = append(opts.ScriptArgs, args[i:]...)
			break
		}

		// Flags --------------------------------------------------------
		var err error
		switch {
		case arg == "-v" || arg == "--version":
			opts.Version = true

		case arg == "-h" || arg == "--help":
			opts.Help = true

		case arg == "-i" || arg == "--interactive":
			opts.Interactive = true

		case arg == "-c" || arg == "--check":
			opts.Check = true

		// --eval / -e
		case arg == "-e" || arg == "--eval":
			opts.Eval, i, err = requireNext(args, i, arg)
		case strings.HasPrefix(arg, "--eval="):
			opts.Eval = arg[len("--eval="):]

		// --print / -p
		case arg == "-p" || arg == "--print":
			opts.Print, i, err = requireNext(args, i, arg)
		case strings.HasPrefix(arg, "--print="):
			opts.Print = arg[len("--print="):]

		// --env-file-if-exists (checked before --env-file to avoid prefix clash)
		case arg == "--env-file-if-exists":
			var v string
			v, i, err = requireNext(args, i, arg)
			opts.EnvFilesIfExist = append(opts.EnvFilesIfExist, v)
		case strings.HasPrefix(arg, "--env-file-if-exists="):
			opts.EnvFilesIfExist = append(opts.EnvFilesIfExist, arg[len("--env-file-if-exists="):])

		// --env-file
		case arg == "--env-file":
			var v string
			v, i, err = requireNext(args, i, arg)
			opts.EnvFiles = append(opts.EnvFiles, v)
		case strings.HasPrefix(arg, "--env-file="):
			opts.EnvFiles = append(opts.EnvFiles, arg[len("--env-file="):])

		default:
			return nil, fmt.Errorf("bad option: %s", arg)
		}

		if err != nil {
			return nil, err
		}
		i++
	}

	return opts, nil
}

// requireNext returns the next argument as the value for flag, advancing i.
func requireNext(args []string, i int, flag string) (string, int, error) {
	i++
	if i >= len(args) {
		return "", i, fmt.Errorf("%s requires an argument", flag)
	}
	return args[i], i, nil
}

// getVersion returns the module version from Go build info when
// available (e.g. installed via `go install`), falling back to a
// development placeholder.
func getVersion() string {
	if info, ok := debug.ReadBuildInfo(); ok {
		v := info.Main.Version
		if v != "" && v != "(devel)" {
			return v
		}
	}
	return "v0.1.0-dev"
}

func printHelp() {
	fmt.Print(`Usage: gode [options] [script.js] [arguments]

Options:
  -v, --version                   print gode version
  -h, --help                      print this help message
  -e, --eval <script>             evaluate script
  -p, --print <script>            evaluate script and print result
  -c, --check                     syntax check script without executing
  -i, --interactive               always enter the REPL even after eval/file
  --env-file <file>               load environment variables from file
  --env-file-if-exists <file>     load env variables from file if it exists
  -                               read script from stdin
  --                              indicate the end of gode options

Environment variables:
  GODE_HISTORY                    path to the REPL history file

Documentation: https://github.com/dop251/goja
`)
}
