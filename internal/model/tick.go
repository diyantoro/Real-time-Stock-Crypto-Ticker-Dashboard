package model

import "time"

type Tick struct {
	Symbol    string
	Price     string
	Volume    string
	Timestamp time.Time
	Source    string
}
