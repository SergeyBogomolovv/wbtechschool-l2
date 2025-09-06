package minishell

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// IsBuiltin проверяет, является ли команда встроенной
func IsBuiltin(name string) bool {
	switch name {
	case "cd", "pwd", "echo", "kill", "ps":
		return true
	default:
		return false
	}
}

// RunBuiltin выполняет встроенную команду
func RunBuiltin(cmd Command, stdin io.Reader, stdout, stderr io.Writer) {
	switch cmd.Args[0] {
	case "cd":
		path := ""
		if len(cmd.Args) > 1 {
			path = cmd.Args[1]
		} else {
			path = os.Getenv("HOME")
		}
		if path == "" {
			fmt.Fprintln(stderr, "cd: no path")
			return
		}
		if !filepath.IsAbs(path) {
			cwd, err := os.Getwd()
			if err != nil {
				fmt.Fprintln(stderr, "cd:", err)
				return
			}
			path = filepath.Join(cwd, path)
		}
		if err := os.Chdir(path); err != nil {
			fmt.Fprintln(stderr, "cd:", err)
		}

	case "pwd":
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(stderr, err)
			return
		}
		fmt.Fprintln(stdout, cwd)

	case "echo":
		fmt.Fprintln(stdout, strings.Join(cmd.Args[1:], " "))

	case "kill":
		if len(cmd.Args) < 2 {
			fmt.Fprintln(stderr, "kill: usage: kill <pid>")
			return
		}
		pid, err := strconv.Atoi(cmd.Args[1])
		if err != nil {
			fmt.Fprintln(stderr, "kill: bad pid")
			return
		}
		proc, err := os.FindProcess(pid)
		if err != nil {
			fmt.Fprintln(stderr, "kill:", err)
			return
		}
		if err := proc.Signal(os.Interrupt); err != nil {
			fmt.Fprintln(stderr, "kill:", err)
		}

	case "ps":
		// ну пока что так
		command := exec.Command("ps", cmd.Args[1:]...)
		command.Stdin = stdin
		command.Stdout = stdout
		command.Stderr = stderr
		if err := command.Run(); err != nil {
			fmt.Fprintln(stderr, err)
		}
	}
}
