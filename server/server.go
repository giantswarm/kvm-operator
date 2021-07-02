// Package server provides a server implementation to connect network transport
// protocols and service business logic by defining server endpoints.
package server

import (
	"context"
	"net/http"
	"sync"

	"github.com/giantswarm/microerror"
	microserver "github.com/giantswarm/microkit/server"
	"github.com/giantswarm/micrologger"
	"github.com/spf13/viper"

	"github.com/giantswarm/kvm-operator/v4/pkg/project"
	"github.com/giantswarm/kvm-operator/v4/server/endpoint"
	"github.com/giantswarm/kvm-operator/v4/service"
)

// Config represents the configuration used to create a new server object.
type Config struct {
	Logger  micrologger.Logger
	Service *service.Service
	Viper   *viper.Viper
}

type Server struct {
	// Dependencies.
	logger micrologger.Logger

	// Internals.
	bootOnce     sync.Once
	config       microserver.Config
	shutdownOnce sync.Once
}

// New creates a new configured server object.
func New(config Config) (*Server, error) {
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Logger must not be empty", config)
	}
	if config.Service == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Service must not be empty", config)
	}
	if config.Viper == nil {
		return nil, microerror.Maskf(invalidConfigError, "%T.Viper must not be empty", config)
	}

	var err error

	var endpointCollection *endpoint.Endpoint
	{
		c := endpoint.Config{
			Logger:  config.Logger,
			Service: config.Service,
		}
		endpointCollection, err = endpoint.New(c)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	s := &Server{
		// Dependencies.
		logger: config.Logger,

		// Internals.
		bootOnce: sync.Once{},
		config: microserver.Config{
			Logger:      config.Logger,
			ServiceName: project.Name(),
			Viper:       config.Viper,

			Endpoints: []microserver.Endpoint{
				endpointCollection.Healthz,
				endpointCollection.Version,
			},
			ErrorEncoder: errorEncoder,
		},
		shutdownOnce: sync.Once{},
	}

	return s, nil
}

func (s *Server) Boot() {
	s.bootOnce.Do(func() {
		// Here goes your custom boot logic for your server/endpoint/middleware, if
		// any.
	})
}

func (s *Server) Config() microserver.Config {
	return s.config
}

func (s *Server) Shutdown() {
	s.shutdownOnce.Do(func() {
		// Here goes your custom shutdown logic for your server/endpoint/middleware,
		// if any.
	})
}

func errorEncoder(ctx context.Context, err error, w http.ResponseWriter) {
	rErr := err.(microserver.ResponseError)
	uErr := rErr.Underlying()

	rErr.SetCode(microserver.CodeInternalError)
	rErr.SetMessage(uErr.Error())
	w.WriteHeader(http.StatusInternalServerError)
}
