package main

import (
	"context"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunSuccess(t *testing.T) {
	cmd := exec.Command("echo")
	err := runWithGracefulShutdown(context.Background(), cmd)
	require.NoError(t, err)
}

func TestRunCancel(t *testing.T) {
	cmd := exec.Command("echo")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := runWithGracefulShutdown(ctx, cmd)
	require.Error(t, err)
	require.ErrorIs(t, err, context.Canceled)
}
