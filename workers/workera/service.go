package workera

import (
	"context"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/woodsaj/go-server/components"
	"github.com/woodsaj/go-server/registry"
)

type WorkerA struct {
	Cfg         *viper.Viper                    `inject:""`
	WorkerPool  *components.WorkerPool          `inject:""`
	PController *components.ProcessorController `inject:""`
}

func init() {
	registry.RegisterService(&WorkerA{}, 9)
	viper.SetDefault("worker-a.enabled", false)
	viper.SetDefault("worker-a.data", "workerA")
}

func (s *WorkerA) Init() error {
	log.Debug("Initializing WorkerA svc")
	if s.Cfg.GetString("worker-a.data") == "" {
		return fmt.Errorf("worker-a.data is not set")
	}
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

	ticker := time.NewTicker(time.Second * 5)

	for {
		select {
		case t := <-ticker.C:
			log.Infof("WorkerA: %s", t.String())
		case <-done:
			log.Info("WorkerA shutting down")
			return nil
		}
	}
	return nil
}
