package cfg

import (
	"sync"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// wrapper for setting default values.
func SetDefault(key string, value interface{}) {
	viper.SetDefault(key, value)
}

type Cfg struct {
	*viper.Viper
	sync.Mutex

	listeners []listener
}

type listener func()

func New(v *viper.Viper) *Cfg {
	return &Cfg{
		listeners: make([]listener, 0),
		Viper:     v,
	}
}

func (c *Cfg) OnChange(l listener) {
	c.Lock()
	c.listeners = append(c.listeners, l)
	c.Unlock()
}

func (c *Cfg) notify() {
	c.Lock()
	for _, l := range c.listeners {
		go l()
	}
	c.Unlock()
}

func (c *Cfg) Watch() {
	c.Viper.WatchConfig()
	c.Viper.OnConfigChange(func(e fsnotify.Event) {
		log.Infof("Config file changed: %s", e.Name)
		c.notify()
	})
}
