package processorbar

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/woodsaj/go-server/processor"
	"github.com/woodsaj/go-server/registry"
)

type ProcessorBar struct {
	Cfg         *viper.Viper          `inject:""`
	PController *processor.Controller `inject:""`

	ready chan struct{}
}

func init() {
	registry.RegisterService(&ProcessorBar{}, 10)
	viper.SetDefault("processor-bar.enabled", false)
	viper.SetDefault("processor-bar.data", "ProcessorBar")
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
