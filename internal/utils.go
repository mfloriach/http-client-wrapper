package internal

import (
	"time"

	"github.com/sony/gobreaker/v2"
)

type Configs func(*configs)

type configs struct {
	Headers        map[string]string
	Retries        int
	Timeout        time.Duration
	HideParams     bool
	Throttle       func()
	CircuitBreaker func(do func() error)
}

func defaultConfigs() configs {
	return configs{
		Headers:    map[string]string{},
		Retries:    0,
		Timeout:    time.Second * 1,
		HideParams: false,
		Throttle:   nil,
	}
}

func AddHeaders(headers map[string]string) Configs {
	return func(c *configs) {
		c.Headers = headers
	}
}

func AddRetries(num int) Configs {
	return func(c *configs) {
		c.Retries = num
	}
}

func AddTimeout(t time.Duration) Configs {
	return func(c *configs) {
		c.Timeout = t
	}
}

func HideParams() Configs {
	return func(c *configs) {
		c.HideParams = true
	}
}

func AddThrottle(numRequestPerSecond int) Configs {
	return func(c *configs) {
		c.Throttle = NewThrottle(numRequestPerSecond)
	}
}

func AddCircuitBreaker(settings gobreaker.Settings) Configs {
	return func(c *configs) {
		c.CircuitBreaker = NewCircuitBreaker(settings)
	}
}
