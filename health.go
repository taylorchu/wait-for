package main

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"time"
)

func checkHealth(ctx context.Context, s string) (bool, error) {
	u, err := url.Parse(s)
	if err != nil {
		return false, err
	}
	switch u.Scheme {
	case "tcp":
		var d net.Dialer
		conn, err := d.DialContext(ctx, u.Scheme, u.Host)
		if err != nil {
			return false, err
		}
		defer conn.Close()
		return true, nil
	}
	return false, nil
}

type HealthGroup struct {
	URLs     []string
	All      bool
	Negate   bool
	Timeout  time.Duration
	Interval time.Duration
	Retry    uint64
}

func (g *HealthGroup) Wait(ctx context.Context) error {
	ticker := time.NewTicker(g.Interval)
	defer ticker.Stop()
	var (
		countOk   uint64
		countFail uint64
	)
	for {
		var (
			countHealthy   uint64
			countUnhealthy uint64
		)
		for _, u := range g.URLs {
			func() {
				hctx, cancel := context.WithTimeout(ctx, g.Timeout)
				defer cancel()

				ok, _ := checkHealth(hctx, u)
				if ok {
					countHealthy += 1
				} else {
					countUnhealthy += 1
				}
			}()
		}
		if g.Negate {
			countHealthy, countUnhealthy = countUnhealthy, countHealthy
		}
		ok := countHealthy == uint64(len(g.URLs))
		if !g.All {
			ok = ok || countHealthy > 0
		}
		if ok {
			countOk += 1
			countFail = 0
		} else {
			countOk = 0
			countFail += 1
		}
		if countOk > 0 {
			return nil
		}
		if g.Retry > 0 && countFail >= g.Retry {
			return fmt.Errorf("wait timeout")
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
