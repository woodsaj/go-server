package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/woodsaj/go-server/cfg"
	"github.com/woodsaj/go-server/components"
	"github.com/woodsaj/go-server/registry"
	"gopkg.in/macaron.v1"
)

func init() {
	registry.RegisterService(&Api{}, 5)
	cfg.SetDefault("api.listen", ":8080")
}

type Api struct {
	Cfg         *cfg.Cfg                        `inject:""`
	WorkerPool  *components.WorkerPool          `inject:""`
	PController *components.ProcessorController `inject:""`

	processor components.Processor
	ctx       context.Context
}

func (a *Api) Init() error {
	log.Debug("Initializing Api service")

	// validate config
	listen := a.Cfg.GetString("api.listen")
	if listen == "" {
		return fmt.Errorf("api.listen is not set")
	}
	host := strings.Split(listen, ":")
	port, err := strconv.ParseInt(host[len(host)-1], 10, 64)
	if err != nil {
		return fmt.Errorf("Could not parse api.listen address. %s", err)
	}
	if port < 0 || port > 65535 {
		return fmt.Errorf("Invalid TCP port for listen address.")
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
	log.Infof("Api server listening on %s", l.Addr().String())
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
