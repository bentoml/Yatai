package utils

import (
	"time"
)

func TimePtr(t time.Time) *time.Time {
	return &t
}

func DurationPtr(d time.Duration) *time.Duration {
	return &d
}

type Waiter interface {
	Wait()
}

func WaitTimeout(waiter Waiter, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		waiter.Wait()
	}()
	select {
	case <-c:
		return false // completed normally
	case <-time.After(timeout):
		return true // timed out
	}
}
