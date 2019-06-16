package worker

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/noaway/juxing/internal/utils"
)

// Mould struct
type Mould func(context.Context, ...interface{}) error

type worker struct {
	name      string
	workerCnt int
	m         Mould
	channel   chan []interface{}
	cancel    context.CancelFunc
	ctx       context.Context
}

func (w *worker) run(num int, fn Mould) {
	for args := range w.channel {
		func() {
			defer utils.DeferError(func(stack string, err interface{}) {
				logrus.Errorf("worker.run.trace err: %v stack info: %v", err, stack)
			})

			for {
				if err := fn(w.ctx, args...); err != nil {
					logrus.Error("worker.run.exec err ", err)
					time.Sleep(time.Second * 2)
					continue
				}
				break
			}
		}()
	}
}

func (w *worker) Close() {
	if w.channel != nil {
		close(w.channel)
	}
	if w.cancel != nil {
		w.cancel()
	}
}

// Workers struct
type Workers struct {
	sync.Map
}

// NewWorkers func
func NewWorkers() *Workers {
	return &Workers{}
}

// Run func
func (ws *Workers) run(w *worker) {
	for i := 0; i < w.workerCnt; i++ {
		go w.run(i, w.m)
	}
}

// HandleFunc func
func (ws *Workers) HandleFunc(name string, m Mould, workerCnt int) bool {
	if _, ok := ws.Load(name); !ok {
		ctx, cancel := context.WithCancel(context.Background())
		w := &worker{
			m:         m,
			ctx:       ctx,
			name:      name,
			cancel:    cancel,
			workerCnt: workerCnt,
			channel:   make(chan []interface{}, workerCnt*2),
		}
		ws.run(w)
		ws.LoadOrStore(name, w)
		return true
	}
	return false
}

// Channels fn
func (ws *Workers) Channels() map[string]*worker {
	ret := map[string]*worker{}
	ws.Range(func(key, value interface{}) bool {
		ret[key.(string)] = value.(*worker)
		return true
	})
	return ret
}

// Remove func
func (ws *Workers) Remove(name string) {
	if w, ok := ws.Load(name); ok {
		w.(*worker).Close()
		ws.Delete(name)
	}
}

// Transmit func
func (ws *Workers) Transmit(name string, args ...interface{}) {
	if w, ok := ws.Load(name); ok {
		w.(*worker).channel <- args
	}
}

// Close func
func (ws *Workers) Close() {
	ws.Range(func(key, value interface{}) bool {
		value.(*worker).Close()
		return true
	})
}

// Go fn
func Go(fn interface{}, args ...interface{}) {
	nf := utils.NewFunction(fn)
	go func() {
		defer utils.DeferError(func(stack string, err interface{}) {
			logrus.Errorf("Go defer err: %v stack info: %v", err, stack)
		})
		nf.Invoke(args...)
	}()
}
