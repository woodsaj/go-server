package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/woodsaj/go-server/processor"
	"github.com/woodsaj/go-server/registry"
	"github.com/woodsaj/go-server/workers"
	"gopkg.in/macaron.v1"
)

func init() {
	registry.RegisterService(&Api{}, 5)
	viper.SetDefault("api.listen", ":8080")
}

type Api struct {
	Cfg         *viper.Viper          `inject:""`
	WorkerPool  *workers.Pool         `inject:""`
	PController *processor.Controller `inject:""`

	processor processor.Processor
	ctx       context.Context
}

func (a *Api) Init() error {
	log.Debug("Initializing Api svc")
	// validate config

	if a.Cfg.GetString("api.listen") == "" {
		return fmt.Errorf("api.listen is not set")
	}
	a.processor = a.PController.Get()
	return nil
}

func (a *Api) Run(ctx context.Context) error {
	a.ctx = ctx
	addr := a.Cfg.GetString("api.listen")
	m := macaron.New()
	m.Use(macaron.Logger())
	m.Use(macaron.Recovery())
	m.Use(macaron.Renderer())
	m.Get("/", a.Hello)
	m.Get("/processor", a.Processor)
	m.Get("/workers", a.Workers)

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	go a.handleShutdown(l)
	srv := http.Server{
		Addr:    addr,
		Handler: m,
	}
	log.Infof("Api server listening on %s", addr)
	err = srv.Serve(l)
	if ctx.Err() != nil {
		return nil
	}
	return err
}

func (a *Api) handleShutdown(l net.Listener) {
	<-a.ctx.Done()
	log.Info("API shutdown started.")
	l.Close()
}

func (a *Api) Hello(ctx *macaron.Context) {
	ctx.PlainText(200, []byte("hello"))
	return
}

func (a *Api) Processor(ctx *macaron.Context) {
	ctx.PlainText(200, []byte(a.processor.Data()))
	return
}

func (a *Api) Workers(ctx *macaron.Context) {
	ctx.PlainText(200, []byte(strings.Join(a.WorkerPool.Status(), "\n")))
	return
}
