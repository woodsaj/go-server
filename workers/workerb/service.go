package workerb

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/woodsaj/go-server/processor"
	"github.com/woodsaj/go-server/registry"
	"github.com/woodsaj/go-server/workers"
)

type WorkerB struct {
	Cfg         *viper.Viper          `inject:""`
	WorkerPool  *workers.Pool         `inject:""`
	PController *processor.Controller `inject:""`
}

func init() {
	registry.RegisterService(&WorkerB{}, 9)
	viper.SetDefault("worker-b.enabled", false)
	viper.SetDefault("worker-b.data", "workerA")
}

func (s *WorkerB) Init() error {
	log.Debug("Initializing WorkerB svc")
	// validate config

	if s.Cfg.GetString("worker-b.data") == "" {
		return fmt.Errorf("worker-b.data is not set")
	}

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

	ticker := time.NewTicker(time.Second * 5)

	for {
		select {
		case t := <-ticker.C:
			log.Infof("WorkerB: %s", t.String())
		case <-done:
			log.Info("WorkerB shutting down")
			return nil
		}
	}
	return nil
}
