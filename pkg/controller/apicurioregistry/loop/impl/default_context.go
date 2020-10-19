package impl

import (
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/loop"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/client"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/configuration"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/env"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/factory"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/patcher"
	"github.com/Apicurio/apicurio-registry-operator/pkg/controller/apicurioregistry/svc/resources"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	sigs_client "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
)

var _ loop.ControlLoopContext = &defaultContext{}

// A long-lived singleton container for shared components
type defaultContext struct {
	appName      string
	appNamespace string
	log          logr.Logger
	requeue      bool
	services     map[string]interface{}
}

// Create a new context when the operator is deployed, provide mostly static data
func NewDefaultContext(appName string, appNamespace string, c controller.Controller, scheme *runtime.Scheme, log logr.Logger, nativeClient sigs_client.Client) *defaultContext {
	this := &defaultContext{
		appName:      appName,
		appNamespace: appNamespace,
		log:          log,
		requeue:      false,
		services:     make(map[string]interface{}, 16),
	}
	this.services[svc.SVC_CONTROLLER] = c
	this.services[svc.SVC_SCHEME] = scheme
	this.services[svc.SVC_NATIVE_CLIENT] = nativeClient

	this.services[svc.SVC_CONFIGURATION] = configuration.NewConfiguration(log)

	this.services[svc.SVC_CLIENTS] = client.NewClients(this)
	this.services[svc.SVC_PATCHERS] = patcher.NewPatchers(this)

	this.services[svc.SVC_KUBE_FACTORY] = factory.NewKubeFactory(this)
	this.services[svc.SVC_OCP_FACTORY] = factory.NewOCPFactory(this)

	this.services[svc.SVC_RESOURCE_CACHE] = resources.NewResourceCache()
	this.services[svc.SVC_ENV_CACHE] = env.NewEnvCache()

	return this
}

func (this *defaultContext) GetLog() logr.Logger {
	return this.log
}

func (this *defaultContext) GetAppName() string {
	return this.appName
}

func (this *defaultContext) GetAppNamespace() string {
	return this.appNamespace
}

func (this *defaultContext) GetService(name string) (interface{}, bool) {
	service, exists := this.services[name]
	return service, exists
}

func (this *defaultContext) RequireService(name string) interface{} {
	service, exists := this.GetService(name)
	if !exists {
		panic("Could not provide service " + name)
	}
	return service
}

func (this *defaultContext) GetController() controller.Controller {
	return this.RequireService(svc.SVC_CONTROLLER).(controller.Controller)
}

func (this *defaultContext) GetScheme() *runtime.Scheme {
	return this.RequireService(svc.SVC_SCHEME).(*runtime.Scheme)
}

func (this *defaultContext) SetRequeue() {
	this.requeue = true
}

func (this *defaultContext) GetAndResetRequeue() bool {
	res := this.requeue
	this.requeue = false
	return res
}