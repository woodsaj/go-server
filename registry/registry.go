package registry

import (
	"context"
	"reflect"
	"sort"

	"github.com/facebookgo/inject"
	log "github.com/sirupsen/logrus"
)

type Descriptor struct {
	Name         string
	Instance     Service
	InitPriority Priority
}

func (d *Descriptor) Inject(serviceGraph *inject.Graph) {
	log.Debugf("adding %s as type %T to dependency graph.", d.Name, d.Instance)
	serviceGraph.Provide(&inject.Object{Value: d.Instance, Name: d.Name})
}

func (d *Descriptor) IsDisabled() bool {
	canBeDisabled, ok := d.Instance.(CanBeDisabled)
	return ok && canBeDisabled.IsDisabled()
}

func (d *Descriptor) BackgroundService() (BackgroundService, bool) {
	svc, ok := d.Instance.(BackgroundService)
	return svc, ok
}

var services []*Descriptor

func RegisterService(instance Service, prio Priority) {
	services = append(services, &Descriptor{
		Name:         reflect.TypeOf(instance).Elem().Name(),
		Instance:     instance,
		InitPriority: prio,
	})
}

func Register(descriptor *Descriptor) {
	services = append(services, descriptor)
}

func GetServices() []*Descriptor {
	sort.Slice(services, func(i, j int) bool {
		return services[i].InitPriority > services[j].InitPriority
	})

	return services
}

// Service interface is the lowest common shape that services
// are expected to forfill to be started within Grafana.
type Service interface {
	// Init is called by at process startup which gives the service
	// the possibility do some initial work before its started. Things
	// like adding routes, bus handlers should be done in the Init function
	Init() error
}

// CanBeDisabled allows the services to decide if it should
// be started or not by itself. This is useful for services
// that might not always be started, ex alerting.
// This will be called after `Init()`.
type CanBeDisabled interface {
	// IsDisabled should return a bool saying if it can be started or not.
	IsDisabled() bool
}

// BackgroundService should be implemented for services that have
// long running tasks in the background.
type BackgroundService interface {
	// Run starts the background process of the service after `Init` have been called
	// on all services. The `context.Context` passed into the function should be used
	// to subscribe to ctx.Done() so the service can be notified when Grafana shuts down.
	Run(ctx context.Context) error
}

type Priority int

const (
	High Priority = 100
	Low  Priority = 0
)
