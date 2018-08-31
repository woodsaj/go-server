package main

import (
	"context"
	"fmt"

	"github.com/facebookgo/inject"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/woodsaj/go-server/registry"
	"golang.org/x/sync/errgroup"
)

type CoreSrv struct {
	context            context.Context
	shutdownFn         context.CancelFunc
	childRoutines      *errgroup.Group
	shutdownReason     string
	shutdownInProgress bool
}

func NewCoreSrv() *CoreSrv {
	rootCtx, shutdownFn := context.WithCancel(context.Background())
	childRoutines, childCtx := errgroup.WithContext(rootCtx)
	return &CoreSrv{
		context:       childCtx,
		shutdownFn:    shutdownFn,
		childRoutines: childRoutines,
	}
}

func (srv *CoreSrv) Run() error {
	serviceGraph := inject.Graph{}

	// inject our config into each service
	// This allows us to just simply provide direct configuration to each service if we dont
	// want to use a configFile, EnvVars or cmdLine args
	serviceGraph.Provide(&inject.Object{Value: viper.GetViper()})

	// inject our logger
	serviceGraph.Provide(&inject.Object{Value: log.StandardLogger()})

	services := registry.GetServices()
	// Add all services to dependency graph
	for _, service := range services {
		serviceGraph.Provide(&inject.Object{Value: service.Instance, Name: service.Name})
	}

	// Inject dependencies to services
	if err := serviceGraph.Populate(); err != nil {
		return fmt.Errorf("Failed to populate service dependency: %v", err)
	}

	// Init & start services
	for _, service := range services {
		if registry.IsDisabled(service.Instance) {
			continue
		}

		log.Info("Initializing " + service.Name)

		if err := service.Instance.Init(); err != nil {
			return fmt.Errorf("Service init failed: %v", err)
		}
	}

	// Start background services
	for _, svc := range services {
		// variable needed for accessing loop variable in function callback
		descriptor := svc
		service, ok := svc.Instance.(registry.BackgroundService)
		if !ok {
			continue
		}

		if registry.IsDisabled(descriptor.Instance) {
			continue
		}

		srv.childRoutines.Go(func() error {
			// Skip starting new service when shutting down
			// Can happen when service stop/return during startup
			if srv.shutdownInProgress {
				return nil
			}

			err := service.Run(srv.context)

			// If error is not canceled then the service crashed
			if err != context.Canceled && err != nil {
				log.Error("Stopped "+descriptor.Name, ". reason: ", err)
			} else {
				log.Info("Stopped "+descriptor.Name, ". reason: ", err)
			}

			// Mark that we are in shutdown mode
			// So more services are not started
			srv.shutdownInProgress = true
			return err
		})
	}

	return srv.childRoutines.Wait()
}

func (srv *CoreSrv) Shutdown(reason string) {
	log.Info("Shutdown started. reason: ", reason)
	srv.shutdownReason = reason
	srv.shutdownInProgress = true

	// call cancel func on root context
	srv.shutdownFn()

	// wait for child routines
	srv.childRoutines.Wait()
}

func (srv *CoreSrv) Exit(reason error) int {
	// default exit code is 1
	code := 1

	if (reason == nil || reason == context.Canceled) && srv.shutdownReason != "" {
		reason = fmt.Errorf(srv.shutdownReason)
		code = 0
	}

	log.Error("Server shutdown. reason: ", reason)
	return code
}
