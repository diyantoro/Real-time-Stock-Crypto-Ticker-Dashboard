package binance

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/diyantoro/ticker-dashboard/internal/model"
)

// Asumsi format Binance (verifikasi ke docs resmi sebelum produksi):
// Combined stream: wss://stream.binance.com:9443/stream?streams=<sym>@trade/...
// Envelope: {"stream":"btcusdt@trade","data":{"e":"trade","E":...,"s":"BTCUSDT","p":"...","q":"...","T":...}}
// Single stream langsung objek trade tanpa envelope.
// Field dipakai: s (symbol), p (price string), q (qty string), T (trade time ms).
// Harga/volume tetap string, tidak pernah float64.

func ParseTrade(raw []byte) (model.Tick, error) {
	payload := raw
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &env); err == nil && len(env.Data) > 0 {
		payload = env.Data
	}
	var tp struct {
		S string    `json:"s"`
		P decString `json:"p"`
		Q decString `json:"q"`
		T int64     `json:"T"`
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(payload, &m); err != nil {
		return model.Tick{}, fmt.Errorf("parse trade: json rusak: %w", err)
	}
	for _, k := range []string{"s", "p", "q", "T"} {
		if _, ok := m[k]; !ok {
			return model.Tick{}, fmt.Errorf("parse trade: field %s hilang", k)
		}
	}
	if err := json.Unmarshal(payload, &tp); err != nil {
		return model.Tick{}, fmt.Errorf("parse trade: %w", err)
	}
	if tp.S == "" {
		return model.Tick{}, fmt.Errorf("parse trade: field s kosong")
	}
	price := string(tp.P)
	qty := string(tp.Q)
	if !isDecimal(price) {
		return model.Tick{}, fmt.Errorf("parse trade: harga tidak valid %q", price)
	}
	if !isDecimal(qty) {
		return model.Tick{}, fmt.Errorf("parse trade: volume tidak valid %q", qty)
	}
	if tp.T <= 0 {
		return model.Tick{}, fmt.Errorf("parse trade: field T tidak valid")
	}
	return model.Tick{
		Symbol:    strings.ToUpper(tp.S),
		Price:     price,
		Volume:    qty,
		Timestamp: time.UnixMilli(tp.T).UTC(),
		Source:    "binance",
	}, nil
}

type decString string

func (d *decString) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		return fmt.Errorf("nilai null")
	}
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*d = decString(s)
		return nil
	}
	*d = decString(strings.TrimSpace(string(b)))
	return nil
}

func isDecimal(s string) bool {
	if s == "" {
		return false
	}
	dots := 0
	digits := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '.' {
			dots++
			if dots > 1 {
				return false
			}
			continue
		}
		if c < '0' || c > '9' {
			return false
		}
		digits++
	}
	return digits > 0
}
