package provider

import (
	"math"
	"math/rand/v2"
	"time"
)

const (
	DefaultInitialInterval = time.Second
	DefaultMaxInterval     = 30 * time.Second
	DefaultMultiplier      = 2.0
	DefaultStableAfter     = 60 * time.Second
)

type Backoff struct {
	Initial    time.Duration
	Max        time.Duration
	Multiplier float64
	RandFloat  func() float64
	attempts   int
}

func NewBackoff() *Backoff {
	return &Backoff{
		Initial:    DefaultInitialInterval,
		Max:        DefaultMaxInterval,
		Multiplier: DefaultMultiplier,
		RandFloat:  rand.Float64,
	}
}

func (b *Backoff) Next() time.Duration {
	base := b.Initial
	max := b.Max
	mult := b.Multiplier
	if base <= 0 {
		base = DefaultInitialInterval
	}
	if max <= 0 {
		max = DefaultMaxInterval
	}
	if mult <= 1 {
		mult = DefaultMultiplier
	}
	exp := float64(base) * math.Pow(mult, float64(b.attempts))
	if exp > float64(max) {
		exp = float64(max)
	}
	b.attempts++
	jitter := 1.0
	if b.RandFloat != nil {
		jitter = 0.5 + b.RandFloat()*0.5
	}
	d := time.Duration(exp * jitter)
	if d > max {
		d = max
	}
	if d < 0 {
		d = 0
	}
	return d
}

func (b *Backoff) Reset() {
	b.attempts = 0
}

func (b *Backoff) Attempts() int {
	return b.attempts
}
