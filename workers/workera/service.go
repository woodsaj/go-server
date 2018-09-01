package workera

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/woodsaj/go-server/cfg"
	"github.com/woodsaj/go-server/components"
	"github.com/woodsaj/go-server/registry"
)

type WorkerA struct {
	Cfg         *cfg.Cfg                        `inject:""`
	WorkerPool  *components.WorkerPool          `inject:""`
	PController *components.ProcessorController `inject:""`

	reload chan struct{}
}

func init() {
	registry.RegisterService(&WorkerA{}, 9)

	// startup settings
	cfg.SetDefault("worker-a.enabled", false)

	// runtime settings
	cfg.SetDefault("worker-a.data", "workerA")
	cfg.SetDefault("worker-a.interval", time.Second*2)
}

func (s *WorkerA) Init() error {
	log.Debug("Initializing WorkerA svc")
	if s.Cfg.GetDuration("worker-a.interval") == time.Duration(0) {
		return fmt.Errorf("worker-a.interval must be > 0")
	}

	s.reload = make(chan struct{})
	s.Cfg.OnChange(func() {
		log.Info("workerA detected config change. Applying changes to runtime settings.")
		s.reload <- struct{}{}
	})
	s.WorkerPool.Register(s)
	return nil
}

func (s *WorkerA) IsDisabled() bool {
	return !s.Cfg.GetBool("worker-a.enabled")
}

func (s *WorkerA) DoWork() {
	fmt.Println(s.Cfg.GetString("worker-a.data"))
}

func (s *WorkerA) Status() string {
	return "workerA running"
}

func (s *WorkerA) Run(ctx context.Context) error {
	done := ctx.Done()
	// wait for our Processor to be ready
	p := s.PController.Get()
	log.Info("WorkerA waiting for processor to be ready.")
	select {
	case <-done:
		log.Info("WorkerA shutting down")
		return nil
	case <-p.Ready():
		log.Info("processor ready, starting up WorkerA")
	}

	interval := s.Cfg.GetDuration("worker-a.interval")
	ticker := time.NewTicker(interval)

	for {
		select {
		case t := <-ticker.C:
			log.Infof("WorkerA: %s", t.String())
		case <-done:
			log.Info("WorkerA shutting down")
			return nil
		case <-s.reload:
			// config reloaded.  Adjust or ticker interval if it has changed.
			newInterval := s.Cfg.GetDuration("worker-a.interval")
			if newInterval == 0 {
				log.Errorf("worker-a.interval has been updated to invalid value. %s", newInterval.String())
				continue
			}
			if interval != newInterval && newInterval != 0 {
				interval = newInterval
				ticker.Stop()
				ticker = time.NewTicker(interval)
			}
		}
	}
	return nil
}
