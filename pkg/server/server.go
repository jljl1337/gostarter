package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/jljl1337/gostarter/pkg/transport"
)

type Named interface {
	Name() string
}

type Lifecycle interface {
	Named
	Start() error
	Stop(context.Context) error
}

type Server struct {
	lifecycleList           []Lifecycle
	apiMux                  *http.ServeMux
	mux                     *http.ServeMux
	apiMiddleware           transport.Middleware
	port                    int
	socketPath              string
	socketPerm              os.FileMode
	httpServer              *http.Server
	gracefulShutdownTimeout time.Duration
}

func NewServer(lifecycleList []Lifecycle, optionList ...Option) (*Server, error) {
	server := &Server{
		lifecycleList: lifecycleList,
		apiMux:        http.NewServeMux(),
		mux:           http.NewServeMux(),
	}

	for _, option := range optionList {
		if err := option(server); err != nil {
			return nil, fmt.Errorf("failed to apply server option: %w", err)
		}
	}

	return server, nil
}
