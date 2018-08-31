package workers

import (
	"sync"

	"github.com/woodsaj/go-server/registry"
)

func init() {
	registry.RegisterService(&Pool{}, 99)
}

func (s *Pool) Init() error {
	s.workers = make([]Worker, 0)
	return nil
}

type Worker interface {
	DoWork()
	Status() string
}

type Pool struct {
	workers []Worker
	sync.Mutex
}

func (wp *Pool) Register(w Worker) {
	wp.Lock()
	wp.workers = append(wp.workers, w)
	wp.Unlock()
}

func (wp *Pool) Status() []string {
	wp.Lock()
	result := make([]string, len(wp.workers))
	for i, w := range wp.workers {
		result[i] = w.Status()
	}
	wp.Unlock()
	return result
}
