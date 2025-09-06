package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"minishell"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	var running atomic.Bool
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT)
	defer signal.Stop(sigch)

	// Перехватываем Ctrl+C и ничего не делаем если не запущен пайплайн
	go func() {
		for range sigch {
			if !running.Load() {
				fmt.Fprintln(os.Stdout)
				fmt.Fprintln(os.Stdout, "minishell>")
			}
		}
	}()

	for {
		fmt.Fprint(os.Stdout, "minishell> ")

		line, err := reader.ReadString('\n')
		// Получили Ctrl+D
		if errors.Is(err, io.EOF) {
			fmt.Println("\nexit")
			return
		}
		if err != nil {
			fmt.Fprintln(os.Stderr, "read error:", err)
			continue
		}

		// Подставляем переменные окружения и убираем лишние пробелы
		line = os.ExpandEnv(strings.TrimSpace(line))
		if line == "" {
			continue
		}

		// Извлекаем пайплайн из строки
		pipeline := minishell.ExtractPipeline(line)
		if len(pipeline) == 0 {
			continue
		}

		running.Store(true)
		minishell.RunPipeline(pipeline)
		running.Store(false)
	}
}
