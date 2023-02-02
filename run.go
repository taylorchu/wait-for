package main

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"syscall"
	"time"
)

func runWithGracefulShutdown(ctx context.Context, cmd *exec.Cmd) error {
	err := cmd.Start()
	if err != nil {
		return err
	}
	errc := make(chan error)
	go func() {
		select {
		case errc <- nil:
			return
		case <-ctx.Done():
		}

		syscall.Kill(cmd.Process.Pid, syscall.SIGINT)

		timer := time.NewTimer(30 * time.Second)
		defer timer.Stop()

		select {
		case errc <- ctx.Err():
			return
		case <-timer.C:
			syscall.Kill(cmd.Process.Pid, syscall.SIGKILL)
		}
		errc <- ctx.Err()
	}()

	waitErr := cmd.Wait()
	if interruptErr := <-errc; interruptErr != nil {
		if errors.Is(interruptErr, context.Canceled) {
			return nil
		}
		return interruptErr
	}
	if waitErr != nil {
		return fmt.Errorf("run command: %w with args %v", waitErr, cmd.Args)
	}
	return nil
}
