package binance

import (
	"context"
	"testing"
	"time"

	"github.com/diyantoro/ticker-dashboard/internal/model"
)

func TestParseTrade(t *testing.T) {
	tests := []struct {
		name   string
		raw    string
		symbol string
		price  string
		volume string
	}{
		{
			name:   "combined stream",
			raw:    `{"stream":"btcusdt@trade","data":{"e":"trade","s":"BTCUSDT","p":"67123.45","q":"0.012","T":1700000000123}}`,
			symbol: "BTCUSDT",
			price:  "67123.45",
			volume: "0.012",
		},
		{
			name:   "single stream numeric decimals",
			raw:    `{"e":"trade","s":"ethusdt","p":2500.5,"q":1,"T":1700000000123}`,
			symbol: "ETHUSDT",
			price:  "2500.5",
			volume: "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTrade([]byte(tt.raw))
			if err != nil {
				t.Fatalf("ParseTrade() error = %v", err)
			}
			if got.Symbol != tt.symbol || got.Price != tt.price || got.Volume != tt.volume {
				t.Errorf("tick = %+v", got)
			}
			if got.Timestamp != time.UnixMilli(1700000000123).UTC() {
				t.Errorf("Timestamp = %v", got.Timestamp)
			}
			if got.Source != "binance" {
				t.Errorf("Source = %q, want binance", got.Source)
			}
		})
	}
}

func TestParseTradeRejectsInvalidPayload(t *testing.T) {
	invalid := []string{
		`{`,
		`{"s":"BTCUSDT","p":"1.2","q":"0.1"}`,
		`{"s":"BTCUSDT","p":"1e3","q":"0.1","T":1700000000123}`,
		`{"s":"BTCUSDT","p":"1.2","q":"-0.1","T":1700000000123}`,
		`{"s":"BTCUSDT","p":"1.2","q":"0.1","T":0}`,
	}
	for _, raw := range invalid {
		t.Run(raw, func(t *testing.T) {
			if _, err := ParseTrade([]byte(raw)); err == nil {
				t.Fatal("ParseTrade() error = nil, want error")
			}
		})
	}
}

func TestBuildStreamURL(t *testing.T) {
	got := BuildStreamURL("wss://example.test/", []string{" BTCUSDT ", "ethusdt", ""})
	want := "wss://example.test/stream?streams=btcusdt@trade/ethusdt@trade"
	if got != want {
		t.Fatalf("BuildStreamURL() = %q, want %q", got, want)
	}
}

func TestRunRejectsEmptySymbols(t *testing.T) {
	client := &Client{}
	err := client.Run(context.Background(), nil, make(chan model.Tick))
	if err == nil {
		t.Fatal("Run() error = nil, want error")
	}
}
