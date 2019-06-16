package worker

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/noaway/juxing/internal/utils"
)

type LineFunc func(context.Context) error

// Guardian struct
type Guardian struct {
	sync.WaitGroup
	sync.Mutex

	gctx    context.Context
	gcannel context.CancelFunc
	timeout time.Duration
}

func NewGuardian() Guardian {
	return newGuardian()
}

func newGuardian() Guardian {
	ctx, cancel := context.WithCancel(context.Background())
	return Guardian{gctx: ctx, gcannel: cancel}
}

// BreadBoard func
func (g *Guardian) BreadBoard(lines ...LineFunc) {
	for _, line := range lines {
		if line == nil {
			continue
		}
		g.Add(1)
		go g.Run(line)
	}
}

// Trace func
func (g *Guardian) Trace(v ...interface{}) {
	logrus.Info(v)
}

// Run func
func (g *Guardian) Run(line LineFunc) {
	for {
		select {
		case <-g.gctx.Done():
			return
		default:
			err := func() error {
				defer utils.DeferError(func(stack string, err interface{}) {
					logrus.Errorf("Guardian.Run.trace err: %v stack info: %v", err, stack)
				})

				if err := line(g.gctx); err != nil {
					logrus.Error("Guardian.return.line err: ", err)
					return err
				}
				return nil
			}()
			if err != nil {
				logrus.Info("2 seconds later retry err ", err)
			}
			time.Sleep(time.Second * 2)
		}
	}
}

// Do factor is called by all delayed tasks
func (g *Guardian) Do(d time.Duration, fn func() error) {
	g.BreadBoard(func(ctx context.Context) error {
		if err := fn(); err != nil {
			return err
		}
		if d == time.Duration(0) {
			return nil
		}
		tick := time.NewTicker(d)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				logrus.Info("sched is closed")
				return nil
			case <-tick.C:
				if err := fn(); err != nil {
					return err
				}
			}
		}
	})
}

func (g *Guardian) Close() {
	g.gcannel()
}
