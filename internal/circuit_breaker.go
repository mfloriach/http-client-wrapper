package internal

import (
	"github.com/sony/gobreaker/v2"
)

type circuitBreaker struct {
	cb *gobreaker.CircuitBreaker[[]byte]
}

func NewCircuitBreaker(settings gobreaker.Settings) func(do func() error) {
	c := circuitBreaker{
		cb: gobreaker.NewCircuitBreaker[[]byte](settings),
	}

	return func(do func() error) {
		c.cb.Execute(func() ([]byte, error) {
			return []byte{}, do()
		})
	}
}
