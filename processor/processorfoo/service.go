package processorfoo

import (
	log "github.com/sirupsen/logrus"
	"github.com/woodsaj/go-server/cfg"
	"github.com/woodsaj/go-server/components"
	"github.com/woodsaj/go-server/registry"
)

type ProcessorFoo struct {
	Cfg         *cfg.Cfg                        `inject:""`
	PController *components.ProcessorController `inject:""`

	ready chan struct{}
}

func init() {
	registry.RegisterService(&ProcessorFoo{}, 10)

	// startup settings
	cfg.SetDefault("processor-foo.enabled", false)

	// runtime settings
	cfg.SetDefault("processor-foo.data", "ProcessorFoo")
}

func (p *ProcessorFoo) Init() error {
	log.Debug("Initializing ProcessorFoo svc")
	p.ready = make(chan struct{})
	err := p.PController.Set(p)
	if err != nil {
		return err
	}
	//imediately ready
	close(p.ready)
	return nil
}

func (p *ProcessorFoo) IsDisabled() bool {
	return !p.Cfg.GetBool("processor-foo.enabled")
}

func (p *ProcessorFoo) Data() string {
	return p.Cfg.GetString("processor-foo.data")
}

func (p *ProcessorFoo) Ready() <-chan struct{} {
	return p.ready
}
