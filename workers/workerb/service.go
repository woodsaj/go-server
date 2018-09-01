package workerb

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/woodsaj/go-server/cfg"
	"github.com/woodsaj/go-server/components"
	"github.com/woodsaj/go-server/registry"
)

type WorkerB struct {
	Cfg         *cfg.Cfg                        `inject:""`
	WorkerPool  *components.WorkerPool          `inject:""`
	PController *components.ProcessorController `inject:""`

	reload chan struct{}
}

func init() {
	registry.RegisterService(&WorkerB{}, 9)

	// startup settings
	cfg.SetDefault("worker-b.enabled", false)

	// runtime settings
	cfg.SetDefault("worker-b.data", "workerA")
	cfg.SetDefault("worker-b.interval", time.Second*1)
}

func (s *WorkerB) Init() error {
	log.Debug("Initializing WorkerB svc")

	// validate config

	if s.Cfg.GetDuration("worker-b.interval") == time.Duration(0) {
		return fmt.Errorf("worker-b.interval must be > 0")
	}

	s.reload = make(chan struct{})
	s.Cfg.OnChange(func() {
		log.Info("workerB detected config change. Applying changes to runtime settings.")
		s.reload <- struct{}{}
	})

	s.WorkerPool.Register(s)
	return nil
}

func (s *WorkerB) IsDisabled() bool {
	return !s.Cfg.GetBool("worker-b.enabled")
}

func (s *WorkerB) DoWork() {
	fmt.Println(s.Cfg.GetString("worker-b.data"))
}

func (s *WorkerB) Status() string {
	return "workerB running"
}

func (s *WorkerB) Run(ctx context.Context) error {
	done := ctx.Done()
	// wait for our Processor to be ready
	p := s.PController.Get()
	log.Info("WorkerB waiting for processor to be ready.")
	select {
	case <-done:
		log.Info("WorkerB shutting down")
		return nil
	case <-p.Ready():
		log.Info("processor ready, starting up WorkerB")
	}

	interval := s.Cfg.GetDuration("worker-b.interval")
	ticker := time.NewTicker(interval)

	for {
		select {
		case t := <-ticker.C:
			log.Infof("WorkerB: %s", t.String())
		case <-done:
			log.Info("WorkerB shutting down")
			return nil
		case <-s.reload:
			// config reloaded.  Adjust or ticker interval if it has changed.
			newInterval := s.Cfg.GetDuration("worker-b.interval")
			if newInterval == 0 {
				log.Errorf("worker-b.interval has been updated to invalid value. %s", newInterval.String())
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
