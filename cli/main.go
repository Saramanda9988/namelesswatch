package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		printRootUsage(stderr)
		return 2
	}

	switch args[0] {
	case "init":
		return runInit(args[1:], stdout, stderr)
	case "validate":
		return runValidate(args[1:], stdout, stderr)
	case "help", "-h", "--help":
		printRootUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n\n", args[0])
		printRootUsage(stderr)
		return 2
	}
}

func runInit(args []string, stdout io.Writer, stderr io.Writer) int {
	opts, target, help, err := parseInitArgs(args)
	if help {
		printInitUsage(stdout)
		return 0
	}
	if err != nil {
		fmt.Fprintf(stderr, "init: %v\n\n", err)
		printInitUsage(stderr)
		return 2
	}

	written, err := ScaffoldPack(target, opts)
	if err != nil {
		fmt.Fprintf(stderr, "init failed: %v\n", err)
		return 1
	}

	absTarget, _ := filepath.Abs(target)
	fmt.Fprintf(stdout, "initialized story pack: %s\n", absTarget)
	fmt.Fprintf(stdout, "written files: %d\n", len(written))
	for _, fileName := range written {
		fmt.Fprintf(stdout, "  - %s\n", fileName)
	}
	return 0
}

func runValidate(args []string, stdout io.Writer, stderr io.Writer) int {
	target, help, err := parseValidateArgs(args)
	if help {
		printValidateUsage(stdout)
		return 0
	}
	if err != nil {
		fmt.Fprintf(stderr, "validate: %v\n\n", err)
		printValidateUsage(stderr)
		return 2
	}

	report, err := ValidatePack(target)
	if err != nil {
		fmt.Fprintf(stderr, "validate failed: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "checked story pack: %s\n", report.Root)
	if report.Title != "" {
		fmt.Fprintf(stdout, "title: %s\n", report.Title)
	}
	if len(report.Problems) == 0 {
		fmt.Fprintln(stdout, "status: ok")
	} else {
		fmt.Fprintln(stdout, "status: invalid")
		fmt.Fprintln(stdout, "problems:")
		for _, problem := range report.Problems {
			fmt.Fprintf(stdout, "  - %s\n", problem)
		}
	}
	if len(report.Warnings) > 0 {
		fmt.Fprintln(stdout, "warnings:")
		for _, warning := range report.Warnings {
			fmt.Fprintf(stdout, "  - %s\n", warning)
		}
	}

	if len(report.Problems) > 0 {
		return 1
	}
	return 0
}

func parseInitArgs(args []string) (ScaffoldOptions, string, bool, error) {
	opts := ScaffoldOptions{}
	target := "."
	var positionals []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-h" || arg == "--help" || arg == "help":
			return opts, "", true, nil
		case arg == "-f" || arg == "--force":
			opts.Force = true
		case arg == "--title" || arg == "-title":
			value, next, err := nextArgValue(args, i, arg)
			if err != nil {
				return opts, "", false, err
			}
			opts.Title = value
			i = next
		case strings.HasPrefix(arg, "--title="):
			opts.Title = strings.TrimPrefix(arg, "--title=")
		case strings.HasPrefix(arg, "-title="):
			opts.Title = strings.TrimPrefix(arg, "-title=")
		case arg == "--initial-scene" || arg == "-initial-scene":
			value, next, err := nextArgValue(args, i, arg)
			if err != nil {
				return opts, "", false, err
			}
			opts.InitialScene = value
			i = next
		case strings.HasPrefix(arg, "--initial-scene="):
			opts.InitialScene = strings.TrimPrefix(arg, "--initial-scene=")
		case strings.HasPrefix(arg, "-initial-scene="):
			opts.InitialScene = strings.TrimPrefix(arg, "-initial-scene=")
		case strings.HasPrefix(arg, "-"):
			return opts, "", false, fmt.Errorf("unknown flag %s", arg)
		default:
			positionals = append(positionals, arg)
		}
	}

	if len(positionals) > 1 {
		return opts, "", false, fmt.Errorf("expected at most one target path, got %d", len(positionals))
	}
	if len(positionals) == 1 {
		target = positionals[0]
	}
	return opts, target, false, nil
}

func parseValidateArgs(args []string) (string, bool, error) {
	target := "."
	var positionals []string

	for _, arg := range args {
		switch {
		case arg == "-h" || arg == "--help" || arg == "help":
			return "", true, nil
		case strings.HasPrefix(arg, "-"):
			return "", false, fmt.Errorf("unknown flag %s", arg)
		default:
			positionals = append(positionals, arg)
		}
	}
	if len(positionals) > 1 {
		return "", false, fmt.Errorf("expected at most one target path, got %d", len(positionals))
	}
	if len(positionals) == 1 {
		target = positionals[0]
	}
	return target, false, nil
}

func nextArgValue(args []string, index int, flagName string) (string, int, error) {
	next := index + 1
	if next >= len(args) {
		return "", index, fmt.Errorf("%s requires a value", flagName)
	}
	if strings.HasPrefix(args[next], "-") {
		return "", index, fmt.Errorf("%s requires a value", flagName)
	}
	return args[next], next, nil
}

func printRootUsage(w io.Writer) {
	fmt.Fprint(w, `Usage:
  namelesswatch-cli init [flags] [path]
  namelesswatch-cli validate [path]

Commands:
  init      Create a story pack scaffold.
  validate  Check story pack files and JSON structure.
`)
}

func printInitUsage(w io.Writer) {
	fmt.Fprint(w, `Usage:
  namelesswatch-cli init [flags] [path]

Flags:
  --title <value>          Story title. Defaults to the target directory name.
  --initial-scene <value>  Initial scene ID. Defaults to "entrance".
  -f, --force              Overwrite existing scaffold files.
`)
}

func printValidateUsage(w io.Writer) {
	fmt.Fprint(w, `Usage:
  namelesswatch-cli validate [path]
`)
}
