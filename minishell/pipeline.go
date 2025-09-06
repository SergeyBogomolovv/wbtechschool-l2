package minishell

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

// Command представляет команду для выполнения
type Command struct {
	Args []string
}

func splitPipes(s string) []string {
	parts := strings.Split(s, "|")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// ExtractPipeline преобразует строку в пайплайн команд
func ExtractPipeline(line string) []Command {
	parts := splitPipes(line)
	pipeline := make([]Command, 0, len(parts))
	for _, p := range parts {
		args := strings.Fields(p)
		if len(args) == 0 {
			continue
		}
		pipeline = append(pipeline, Command{Args: args})
	}
	return pipeline
}

// RunPipeline выполняет пайплайн команд
func RunPipeline(pipeline []Command, stdin io.Reader, stdout, stderr io.Writer) {
	if len(pipeline) == 0 {
		return
	}

	// Одиночный builtin выполняем в родительском процессе
	if len(pipeline) == 1 && IsBuiltin(pipeline[0].Args[0]) {
		RunBuiltin(pipeline[0], stdin, stdout, stderr)
		return
	}

	// Запретим cd в конвейере, чтобы не мутировать состояние shell в середине пайплайна
	for _, cmd := range pipeline {
		if cmd.Args[0] == "cd" {
			fmt.Fprintln(stderr, "cd in pipeline is not supported")
			return
		}
	}

	// Канал для перехвата SIGINT во время выполнения пайплайна
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	defer signal.Stop(sigch)

	var (
		externals  []*exec.Cmd
		leaderPID  int
		prevRead   io.ReadCloser
		wgBuiltins sync.WaitGroup
	)

	// Отдельная горутина, чтобы по Ctrl+C послать сигнал всей группе
	done := make(chan struct{})
	defer close(done)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-sigch:
				if leaderPID != 0 {
					// Отправим SIGINT всей процесс-группе пайплайна
					syscall.Kill(-leaderPID, syscall.SIGINT)
					// На всякий случай продублируем в каждый известный процесс
					for _, cmd := range externals {
						if cmd != nil && cmd.Process != nil {
							cmd.Process.Signal(os.Interrupt)
						}
					}
				}
			}
		}
	}()

	for i, cmd := range pipeline {
		var (
			inR  io.ReadCloser
			outW io.WriteCloser
		)

		if prevRead != nil {
			inR = prevRead
		}

		// Для всех, кроме последнего, создаём промежуточный pipe
		if i < len(pipeline)-1 {
			rp, wp, err := os.Pipe()
			if err != nil {
				fmt.Fprintln(stderr, err)
				return
			}
			outW = wp
			prevRead = rp
		} else {
			prevRead = nil
		}

		var in io.Reader = stdin
		var out io.Writer = stdout
		if inR != nil {
			in = inR
		}
		if outW != nil {
			out = outW
		}

		if IsBuiltin(cmd.Args[0]) {
			// builtin внутри пайплайна исполним в горутине
			wgBuiltins.Go(func() {
				RunBuiltin(cmd, in, out, stderr)
				// Закрыть концы трубы, чтобы downstream получил EOF
				if inR != nil {
					inR.Close()
				}
				if outW != nil {
					outW.Close()
				}
			})
			continue
		}

		// Внешняя команда
		ecmd := exec.Command(cmd.Args[0], cmd.Args[1:]...)
		ecmd.Stdin = in
		ecmd.Stdout = out
		ecmd.Stderr = stderr

		// Объединяем все процессы пайплайна в одну группу
		if leaderPID == 0 {
			ecmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		} else {
			ecmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pgid: leaderPID}
		}

		if err := ecmd.Start(); err != nil {
			fmt.Fprintln(stderr, err)
			// Закроем неиспользуемые дескрипторы
			if inR != nil {
				inR.Close()
			}
			if outW != nil {
				outW.Close()
			}
			return
		}

		externals = append(externals, ecmd)

		if leaderPID == 0 && ecmd.Process != nil {
			leaderPID = ecmd.Process.Pid
		}

		// Родителю эти концы больше не нужны
		if inR != nil {
			inR.Close()
		}
		if outW != nil {
			outW.Close()
		}
	}

	// Дождаться внешних процессов
	for _, cmd := range externals {
		if err := cmd.Wait(); err != nil {
			fmt.Fprintln(stderr, err)
		}
	}

	// Дождаться builtin-ов
	wgBuiltins.Wait()
}
