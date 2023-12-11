package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"strings"
)

func main() {
	exit := make(chan struct{}, 2) // buffer this so there's no deadlock.
	runLoop(os.Stdin, os.Stdout, os.Stderr, exit)
}

func runLoop(r io.Reader, w, errW io.Writer, exit chan struct{}) {
	var (
		input    string
		err      error
		readLoop = bufio.NewReader(r)
	)
	for {
		select {
		case <-exit:
			_, _ = fmt.Fprintln(w, "exiting gracefully...")
			return
		default:
			if err := printPrompt(w); err != nil {
				_, _ = fmt.Fprintln(errW, err)
				continue
			}
			if input, err = readLoop.ReadString('\n'); err != nil {
				_, _ = fmt.Fprintln(errW, err)
				continue
			}
			if err = handleInput(w, input, exit); err != nil {
				_, _ = fmt.Fprintln(errW, err)
			}
		}
	}
}

func printPrompt(w io.Writer) error {
	// Get current user.
	// Don't prematurely memoize this because it might change due to `su`?
	u, err := user.Current()
	if err != nil {
		return err
	}
	// Get current working directory.
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	// /home/User [Username] $
	_, err = fmt.Fprintf(w, "%v [%v] $ ", wd, u.Username)

	return err
}

func handleInput(w io.Writer, input string, exit chan<- struct{}) error {
	// Remove trailing spaces.
	input = strings.TrimSpace(input)

	// Split the input separate the command name and the command arguments.
	args := strings.Fields(input)

	if len(args) == 0 {
		// Empty input, ignore.
		return nil
	}

	name, args := args[0], args[1:]

	// Check for built-in commands.
	switch name {
	case "cd":
		return ChangeDirectory(args...)
	case "env":
		return EnvironmentVariables(w, args...)
	case "exit":
		exit <- struct{}{}
		return nil
	case "echo":
		return Echo(w, args...)
	case "pwd":
		return PrintWorkingDirectory(w)
	case "history":
		return History(w, args...)
	case "mkdir":
		return MakeDirectory(args...)
	case "rm":
		return RemoveFile(args...)
	}

	return executeCommand(name, args...)
}

func ChangeDirectory(args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("cd: missing argument")
	}
	return os.Chdir(args[0])
}

func EnvironmentVariables(w io.Writer, args ...string) error {
	if len(args) == 0 {
		// Print all environment variables if no arguments provided.
		for _, envVar := range os.Environ() {
			_, _ = fmt.Fprintln(w, envVar)
		}
		return nil
	}
	// Otherwise, print the value of the specified environment variable.
	for _, arg := range args {
		value, exists := os.LookupEnv(arg)
		if exists {
			_, _ = fmt.Fprintf(w, "%s=%s\n", arg, value)
		} else {
			_, _ = fmt.Fprintf(w, "%s not found in environment variables\n", arg)
		}
	}
	return nil
}

func Echo(w io.Writer, args ...string) error {
	_, _ = fmt.Fprintln(w, strings.Join(args, " "))
	return nil
}

func PrintWorkingDirectory(w io.Writer) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(w, wd)
	return nil
}

func History(w io.Writer, args ...string) error {
	// Placeholder for history command implementation.
	_, _ = fmt.Fprintln(w, "History command is not implemented yet.")
	return nil
}

func MakeDirectory(args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("mkdir: missing operand")
	}
	for _, dir := range args {
		err := os.Mkdir(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func RemoveFile(args ...string) error {
	if len(args) == 0 {
		return fmt.Errorf("rm: missing operand")
	}
	for _, file := range args {
		err := os.Remove(file)
		if err != nil {
			return err
		}
	}
	return nil
}

func executeCommand(name string, arg ...string) error {
	// Otherwise prep the command
	cmd := exec.Command(name, arg...)

	// Set the correct output device.
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	// Execute the command and return the error.
	return cmd.Run()
}
