package processorbar

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/woodsaj/go-server/cfg"
	"github.com/woodsaj/go-server/components"
	"github.com/woodsaj/go-server/registry"
)

type ProcessorBar struct {
	Cfg         *cfg.Cfg                        `inject:""`
	PController *components.ProcessorController `inject:""`

	ready chan struct{}
}

func init() {
	registry.RegisterService(&ProcessorBar{}, 10)

	//startup settings
	cfg.SetDefault("processor-bar.enabled", false)
	cfg.SetDefault("processor-bar.max-start-delay", time.Second*10)

	// runtime settings
	cfg.SetDefault("processor-bar.data", "ProcessorBar")
}

func (p *ProcessorBar) Init() error {
	log.Debug("Initializing ProcessorBar svc")
	p.ready = make(chan struct{})
	err := p.PController.Set(p)
	if err != nil {
		return err
	}
	return nil
}

func (p *ProcessorBar) IsDisabled() bool {
	return !p.Cfg.GetBool("processor-bar.enabled")
}

func (p *ProcessorBar) Data() string {
	return p.Cfg.GetString("processor-bar.data")
}

func (p *ProcessorBar) Ready() <-chan struct{} {
	return p.ready
}

func (p *ProcessorBar) Run(ctx context.Context) error {
	// simulate a 30second startup time.
	timer := time.NewTimer(time.Second * 30)
	done := ctx.Done()
	select {
	case <-timer.C:
		close(p.ready)
	case <-done:
		return nil
	}
	<-done
	return nil
}
