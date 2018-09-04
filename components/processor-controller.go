package components

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/woodsaj/go-server/registry"
)

func init() {
	registry.RegisterService(&ProcessorController{}, 99)
}

type Processor interface {
	Data() string
	Ready() <-chan struct{}
}

func (c *ProcessorController) Init() error {
	return nil
}

type ProcessorController struct {
	Processor Processor
}

func (c *ProcessorController) Set(p Processor) error {
	log.Infof("setting processor to %T", p)
	if c.Processor != nil {
		return fmt.Errorf("Only 1 processor can be set.")
	}
	c.Processor = p
	return nil
}

func (c *ProcessorController) Get() Processor {
	return c.Processor
}
