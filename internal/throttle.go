package internal

import (
	"sync"

	"go.uber.org/ratelimit"
)

type throttle struct {
	mx sync.Mutex
	rl ratelimit.Limiter
}

func NewThrottle(numRequestPerSecond int) func() {
	t := throttle{
		rl: ratelimit.New(numRequestPerSecond),
		mx: sync.Mutex{},
	}

	return func() {
		t.mx.Lock()
		for {
			t.rl.Take()
			break
		}
		t.mx.Unlock()
	}
}
