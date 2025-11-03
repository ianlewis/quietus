// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// command quietus is a CLI tool to send signals to applications. It can be used
// to gracefully terminate processes by sending a `SIGTERM` then `SIGKILL` after
// a timeout if a process hasn't terminated on its own.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"sync"
	"syscall"
	"time"
)

func waitOnProcess(ctx context.Context, p *os.Process) error {
	waitInterval := 10 * time.Millisecond
	maxWait := 1 * time.Second

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		if err := p.Signal(syscall.Signal(0)); err != nil {
			// Process has exited.
			return nil
		} else {
			time.Sleep(waitInterval)
			waitInterval = min(time.Duration(float64(waitInterval)*1.5), maxWait)
		}
	}
}

func sendAfter(ctx context.Context, p *os.Process, sig syscall.Signal, d time.Duration) error {
	select {
	case <-ctx.Done():
		return nil
	case <-time.After(d):
		return p.Signal(sig)
	}
}

func run(pid int, signal syscall.Signal, timeout time.Duration, secondSignal syscall.Signal) error {
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}

	ctx, cancel := context.WithCancel(context.Background())

	ctx2, cancel2 := context.WithTimeout(ctx, timeout+10*time.Second)
	defer cancel2()

	errChan := make(chan error, 2)

	wg.Add(1)

	go func() {
		defer cancel2()
		defer wg.Done()

		if err := sendAfter(ctx, p, signal, 0); err != nil {
			errChan <- fmt.Errorf("error sending %s: %w", signal, err)
			return
		}

		if secondSignal != 0 {
			if err := sendAfter(ctx, p, secondSignal, timeout); err != nil {
				errChan <- fmt.Errorf("error sending %s: %w", secondSignal, err)
			}
		}
	}()

	wg.Add(1)

	go func() {
		defer cancel()
		defer wg.Done()

		if err := waitOnProcess(ctx2, p); err != nil {
			errChan <- fmt.Errorf("error waiting on PID %d: %w", pid, err)
		}
	}()

	wg.Wait()

	// Return the first error, if any.
	select {
	case err := <-errChan:
		return err
	default:
	}

	return nil
}

func main() {
	signals := map[string]syscall.Signal{
		"SIGABRT":   syscall.SIGABRT,
		"SIGALRM":   syscall.SIGALRM,
		"SIGBUS":    syscall.SIGBUS,
		"SIGCHLD":   syscall.SIGCHLD,
		"SIGCLD":    syscall.SIGCLD,
		"SIGCONT":   syscall.SIGCONT,
		"SIGFPE":    syscall.SIGFPE,
		"SIGHUP":    syscall.SIGHUP,
		"SIGILL":    syscall.SIGILL,
		"SIGINT":    syscall.SIGINT,
		"SIGIO":     syscall.SIGIO,
		"SIGIOT":    syscall.SIGIOT,
		"SIGKILL":   syscall.SIGKILL,
		"SIGPIPE":   syscall.SIGPIPE,
		"SIGPOLL":   syscall.SIGPOLL,
		"SIGPROF":   syscall.SIGPROF,
		"SIGPWR":    syscall.SIGPWR,
		"SIGQUIT":   syscall.SIGQUIT,
		"SIGSEGV":   syscall.SIGSEGV,
		"SIGSTKFLT": syscall.SIGSTKFLT,
		"SIGSTOP":   syscall.SIGSTOP,
		"SIGSYS":    syscall.SIGSYS,
		"SIGTERM":   syscall.SIGTERM,
		"SIGTRAP":   syscall.SIGTRAP,
		"SIGTSTP":   syscall.SIGTSTP,
		"SIGTTIN":   syscall.SIGTTIN,
		"SIGTTOU":   syscall.SIGTTOU,
		"SIGUNUSED": syscall.SIGUNUSED,
		"SIGURG":    syscall.SIGURG,
		"SIGUSR1":   syscall.SIGUSR1,
		"SIGUSR2":   syscall.SIGUSR2,
		"SIGVTALRM": syscall.SIGVTALRM,
		"SIGWINCH":  syscall.SIGWINCH,
		"SIGXCPU":   syscall.SIGXCPU,
		"SIGXFSZ":   syscall.SIGXFSZ,
	}

	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	signal := flagSet.String("signal", "SIGTERM", "signal to send to the process")
	timeout := flagSet.Duration("timeout", 0, "time to wait for process to exit after sending first signal")
	secondSignal := flagSet.String("second-signal", "", "signal to send if process does not exit before wait-timeout (e.g., SIGKILL)")
	listSignals := flagSet.Bool("list", false, "list available signals")

	flagSet.Parse(os.Args[1:])

	if *listSignals {
		for name := range signals {
			fmt.Println(name)
		}
		return
	}

	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: quietus <pid>\n")
		os.Exit(1)
	}

	pid, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid pid: %v\n", err)
		os.Exit(1)
	}

	sig, ok := signals[*signal]
	if !ok {
		fmt.Fprintf(os.Stderr, "invalid signal: %s\n", *signal)
		os.Exit(1)
	}

	var secondSig syscall.Signal
	if *secondSignal != "" {
		secondSig, ok = signals[*secondSignal]
		if !ok {
			fmt.Fprintf(os.Stderr, "invalid second signal: %s\n", *secondSignal)
			os.Exit(1)
		}
	}

	if err := run(pid, sig, *timeout, secondSig); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
