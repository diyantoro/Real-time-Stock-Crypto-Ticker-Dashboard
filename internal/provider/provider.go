package provider

import (
	"context"

	"github.com/diyantoro/ticker-dashboard/internal/model"
)

type Provider interface {
	Name() string
	Run(ctx context.Context, symbols []string, out chan<- model.Tick) error
}
