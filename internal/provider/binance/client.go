package binance

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/diyantoro/ticker-dashboard/internal/model"
	"github.com/diyantoro/ticker-dashboard/internal/provider"
)

type Client struct {
	BaseURL     string
	Logger      *slog.Logger
	NewBackoff  func() *provider.Backoff
	StableAfter time.Duration
	DialTimeout time.Duration
}

func (c *Client) Name() string { return "binance" }

func (c *Client) log() *slog.Logger {
	if c.Logger != nil {
		return c.Logger
	}
	return slog.Default()
}

func BuildStreamURL(base string, symbols []string) string {
	base = strings.TrimRight(base, "/")
	streams := make([]string, 0, len(symbols))
	for _, s := range symbols {
		s = strings.ToLower(strings.TrimSpace(s))
		if s != "" {
			streams = append(streams, s+"@trade")
		}
	}
	return base + "/stream?streams=" + strings.Join(streams, "/")
}

func (c *Client) Run(ctx context.Context, symbols []string, out chan<- model.Tick) error {
	if len(symbols) == 0 {
		return fmt.Errorf("binance: symbols kosong")
	}
	allowed := make(map[string]struct{}, len(symbols))
	for _, s := range symbols {
		allowed[strings.ToUpper(s)] = struct{}{}
	}
	base := c.BaseURL
	if base == "" {
		base = "wss://stream.binance.com:9443"
	}
	url := BuildStreamURL(base, symbols)
	stableAfter := c.StableAfter
	if stableAfter <= 0 {
		stableAfter = provider.DefaultStableAfter
	}
	newBackoff := c.NewBackoff
	if newBackoff == nil {
		newBackoff = provider.NewBackoff
	}
	bo := newBackoff()
	dialTimeout := c.DialTimeout
	if dialTimeout <= 0 {
		dialTimeout = 10 * time.Second
	}
	logger := c.log()

	for {
		if err := ctx.Err(); err != nil {
			return nil
		}
		stable, err := c.connectAndStream(ctx, url, allowed, out, dialTimeout, stableAfter, logger)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			logger.Warn("binance stream berhenti, reconnect", "error", err)
		} else {
			logger.Info("binance stream berhenti bersih")
		}
		if stable {
			bo.Reset()
		}
		d := bo.Next()
		logger.Info("binance reconnect backoff", "delay", d.String())
		t := time.NewTimer(d)
		select {
		case <-ctx.Done():
			t.Stop()
			return nil
		case <-t.C:
		}
	}
}

func (c *Client) connectAndStream(ctx context.Context, url string, allowed map[string]struct{}, out chan<- model.Tick, dialTimeout, stableAfter time.Duration, logger *slog.Logger) (stable bool, err error) {
	dialCtx, cancel := context.WithTimeout(ctx, dialTimeout)
	defer cancel()
	dialer := websocket.DefaultDialer
	conn, _, err := dialer.DialContext(dialCtx, url, nil)
	if err != nil {
		return false, fmt.Errorf("binance dial: %w", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	start := time.Now()

	ctx, stop := context.WithCancel(ctx)
	defer stop()
	go func() {
		<-ctx.Done()
		_ = conn.UnderlyingConn().SetReadDeadline(time.Now())
		_ = conn.Close()
	}()

	const pongWait = 60 * time.Second
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if ctx.Err() != nil {
				return time.Since(start) >= stableAfter, nil
			}
			return time.Since(start) >= stableAfter, fmt.Errorf("binance read: %w", err)
		}
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))
		t, err := ParseTrade(msg)
		if err != nil {
			logger.Warn("pesan rusak, abaikan", "error", err)
			continue
		}
		if _, ok := allowed[t.Symbol]; !ok {
			logger.Warn("simbol tak dikenal, abaikan", "symbol", t.Symbol)
			continue
		}
		select {
		case <-ctx.Done():
			return time.Since(start) >= stableAfter, nil
		case out <- t:
		}
	}
}
