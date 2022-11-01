package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCheckHealth(t *testing.T) {
	ok, err := checkHealth(context.Background(), "tcp://google.com:80")
	require.NoError(t, err)
	require.True(t, ok)
}

func TestCheckHealthLocal(t *testing.T) {
	server := http.Server{
		Addr:    ":8888",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
	}
	go server.ListenAndServe()
	defer server.Close()

	time.Sleep(time.Second)

	ok, err := checkHealth(context.Background(), "tcp://:8888")
	require.NoError(t, err)
	require.True(t, ok)
}

func TestHealthGroupAllHealthy(t *testing.T) {
	g := HealthGroup{
		URLs:     []string{"tcp://google.com:80", "tcp://google.com:443"},
		All:      true,
		Negate:   false,
		Retry:    0,
		Interval: time.Second,
		Timeout:  time.Second,
	}
	err := g.Wait(context.Background())
	require.NoError(t, err)
}

func TestHealthGroupSomeHealthy(t *testing.T) {
	g := HealthGroup{
		URLs:     []string{"tcp://google.com:80", "tcp://google.com:81"},
		All:      false,
		Negate:   false,
		Retry:    0,
		Interval: time.Second,
		Timeout:  time.Second,
	}
	err := g.Wait(context.Background())
	require.NoError(t, err)
}

func TestHealthGroupAllUnhealthy(t *testing.T) {
	g := HealthGroup{
		URLs:     []string{"tcp://google.com:81", "tcp://google.com:82"},
		All:      true,
		Negate:   true,
		Retry:    0,
		Interval: time.Second,
		Timeout:  time.Second,
	}
	err := g.Wait(context.Background())
	require.NoError(t, err)
}

func TestHealthGroupSomeUnhealthy(t *testing.T) {
	g := HealthGroup{
		URLs:     []string{"tcp://google.com:80", "tcp://google.com:81"},
		All:      false,
		Negate:   true,
		Retry:    0,
		Interval: time.Second,
		Timeout:  time.Second,
	}
	err := g.Wait(context.Background())
	require.NoError(t, err)
}

func TestHealthGroupRetry(t *testing.T) {
	g := HealthGroup{
		URLs:     []string{"tcp://google.com:81"},
		All:      true,
		Negate:   false,
		Retry:    1,
		Interval: time.Second,
		Timeout:  time.Second,
	}
	err := g.Wait(context.Background())
	require.Error(t, err)
	require.ErrorContains(t, err, "wait timeout")
}

func TestHealthGroupCancel(t *testing.T) {
	g := HealthGroup{
		URLs:     []string{"tcp://google.com:81"},
		All:      true,
		Negate:   false,
		Retry:    0,
		Interval: time.Second,
		Timeout:  time.Second,
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := g.Wait(ctx)
	require.Error(t, err)
	require.ErrorIs(t, err, context.Canceled)
}
