package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	var (
		flagWaitStartURLs     string
		flagWaitStartInterval time.Duration
		flagWaitStartTimeout  time.Duration
		flagWaitStartRetry    uint64

		flagWaitStopURLs     string
		flagWaitStopInterval time.Duration
		flagWaitStopTimeout  time.Duration
		flagWaitStopRetry    uint64
	)

	flag.StringVar(&flagWaitStartURLs, "wait-start", "", "wait for `url` to become healthy to start")
	flag.DurationVar(&flagWaitStartInterval, "wait-start-interval", time.Second, "interval for wait-start")
	flag.DurationVar(&flagWaitStartTimeout, "wait-start-timeout", time.Second, "timeout for wait-start")
	flag.Uint64Var(&flagWaitStartRetry, "wait-start-retry", 10, "retry count for wait-start")

	flag.StringVar(&flagWaitStopURLs, "wait-stop", "", "wait for `url` to become unhealthy to stop")
	flag.DurationVar(&flagWaitStopInterval, "wait-stop-interval", time.Second, "interval for wait-stop")
	flag.DurationVar(&flagWaitStopTimeout, "wait-stop-timeout", time.Second, "timeout for wait-stop")
	flag.Uint64Var(&flagWaitStopRetry, "wait-stop-retry", 10, "retry count for wait-stop")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
	}()

	if flagWaitStartURLs != "" {
		// wait for start urls before we start commands
		g := HealthGroup{
			URLs:     strings.Split(flagWaitStartURLs, ","),
			All:      true,
			Negate:   false,
			Retry:    flagWaitStartRetry,
			Interval: flagWaitStartInterval,
			Timeout:  flagWaitStartTimeout,
		}
		err := g.Wait(ctx)
		if err != nil {
			log.Fatalf("wait start failed: %s\n", err)
		}
	}

	args := flag.Args()
	if len(args) == 0 {
		log.Fatalf("no command to run\n")
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if flagWaitStopURLs != "" {
		go func() {
			defer cancel()

			for _, g := range []HealthGroup{
				// wait for stop urls to be healthy once
				{
					URLs:     strings.Split(flagWaitStopURLs, ","),
					All:      true,
					Negate:   false,
					Retry:    flagWaitStopRetry,
					Interval: flagWaitStopInterval,
					Timeout:  flagWaitStopTimeout,
				},
				// then wait for stop urls to be unhealthy
				{
					URLs:     strings.Split(flagWaitStopURLs, ","),
					All:      false,
					Negate:   true,
					Retry:    0,
					Interval: flagWaitStopInterval,
					Timeout:  flagWaitStopTimeout,
				},
			} {
				err := g.Wait(ctx)
				if err != nil {
					log.Printf("wait stop failed: %s\n", err)
					return
				}
			}
		}()
	}

	err := runWithGracefulShutdown(ctx, cmd)
	if err != nil {
		log.Fatalf("command exited with error: %s\n", err)
	}
}
