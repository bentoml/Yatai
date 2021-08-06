package errsgroup

import (
	"strings"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"

	"github.com/bentoml/yatai/common/consts"

	"github.com/pkg/errors"
)

type Group struct {
	errs     []error
	wg       sync.WaitGroup
	locker   sync.Mutex
	poolSize int
	tasks    []func()
}

func (g *Group) SetPoolSize(size int) {
	g.locker.Lock()
	g.poolSize = size
	g.locker.Unlock()
}

func (g *Group) Go(f func() error) {
	g.wg.Add(1)
	task := func() {
		defer g.wg.Done()
		err := f()
		if err != nil {
			g.locker.Lock()
			defer g.locker.Unlock()
			g.errs = append(g.errs, err)
		}
	}
	g.locker.Lock()
	defer g.locker.Unlock()
	if g.poolSize > 0 {
		g.tasks = append(g.tasks, task)
	} else {
		go task()
	}
}

func (g *Group) Wait() error {
	if g.poolSize > 0 {
		pool, err := ants.NewPool(g.poolSize)
		if err != nil {
			return err
		}
		defer pool.Release()
		for _, t := range g.tasks {
			err := pool.Submit(t)
			if err != nil {
				return err
			}
		}
	}
	g.wg.Wait()
	if len(g.errs) == 0 {
		return nil
	}
	errMsgs := make([]string, 0, len(g.errs))
	for _, err := range g.errs {
		errMsgs = append(errMsgs, err.Error())
	}
	return errors.New(strings.Join(errMsgs, "; "))
}

func (g *Group) WaitWithTimeout(timeout time.Duration) error {
	c := make(chan error)
	go func() {
		c <- g.Wait()
	}()
	select {
	case r := <-c:
		return r
	case <-time.After(timeout):
		return consts.ErrTimeout
	}
}
