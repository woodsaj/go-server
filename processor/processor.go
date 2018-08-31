package processor

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/woodsaj/go-server/registry"
)

func init() {
	registry.RegisterService(&Controller{}, 99)
}

type Processor interface {
	Data() string
	Ready() <-chan struct{}
}

func (c *Controller) Init() error {
	return nil
}

type Controller struct {
	Processor Processor
}

func (c *Controller) Set(p Processor) error {
	log.Infof("setting processor to %s", p.Data())
	if c.Processor != nil {
		return fmt.Errorf("Only 1 processor can be set.")
	}
	c.Processor = p
	return nil
}

func (c *Controller) Get() Processor {
	return c.Processor
}
