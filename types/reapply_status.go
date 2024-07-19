package types

import "time"

type ReapplyStatus struct {
	Count int
	Next  time.Time
}
