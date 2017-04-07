package mass

import (
	"sync"
	"time"
)

type Limiter struct {
	mu         sync.Mutex
	rate       int
	lastAction time.Time
	interval   time.Duration
}

func NewLimiter(rate int) *Limiter {
	return &Limiter{
		rate:     rate,
		interval: time.Microsecond * time.Duration(1000*1000/rate),
	}
}

func (l *Limiter) Limit() bool {
	res := false
	for {
		l.mu.Lock()

		if time.Now().Sub(l.lastAction) >= l.interval {
			l.lastAction = time.Now()
			res = true
		}

		l.mu.Unlock()

		if res {
			return res
		}

		time.Sleep(l.interval)
	}
}
