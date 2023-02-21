package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/repoerna/go-service3/cmd/services/sales-api/handlers"
	"github.com/repoerna/go-service3/pkg/conf"
	"github.com/repoerna/go-service3/pkg/logger"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
)

var build = "develop"

func main() {

	// construct app logger
	log, err := logger.New("SALES_API")
	if err != nil {
		fmt.Println("error creating logger: ", err)
	}

	if err = run(log); err != nil {
		log.Errorw("startup", "ERROR", err)
	}

	// log.Info("stopping service")

}

func run(log *zap.SugaredLogger) error {

	// =========================================================================
	// Configuration

	cfg := struct {
		conf.Version
		Web struct {
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:10s"`
			IdleTimeout     time.Duration `conf:"default:120s"`
			ShutdownTimeout time.Duration `conf:"default:20s"`
			APIHost         string        `conf:"default:0.0.0.0:3000"`
			DebugHost       string        `conf:"default:0.0.0.0:4000"`
		}
		// Auth struct {
		// 	KeysFolder string `conf:"default:zarf/keys/"`
		// 	ActiveKID  string `conf:"default:54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"`
		// }
		// Vault struct {
		// 	Address   string `conf:"default:http://vault-service.sales-system.svc.cluster.local:8200"`
		// 	MountPath string `conf:"default:secret"`
		// 	Token     string `conf:"default:mytoken,mask"`
		// }
		// DB struct {
		// 	User         string `conf:"default:postgres"`
		// 	Password     string `conf:"default:postgres,mask"`
		// 	Host         string `conf:"default:database-service.sales-system.svc.cluster.local"`
		// 	Name         string `conf:"default:postgres"`
		// 	MaxIdleConns int    `conf:"default:2"`
		// 	MaxOpenConns int    `conf:"default:0"`
		// 	DisableTLS   bool   `conf:"default:true"`
		// }
		// Zipkin struct {
		// 	ReporterURI string  `conf:"default:http://zipkin-service.sales-system.svc.cluster.local:9411/api/v2/spans"`
		// 	ServiceName string  `conf:"default:sales-api"`
		// 	Probability float64 `conf:"default:0.05"`
		// }
	}{
		Version: conf.Version{
			Build: build,
			Desc:  "copyright information here",
		},
	}

	const prefix = "SALES"
	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}
		return fmt.Errorf("parsing config: %w", err)
	}

	// =========================================================================
	// GOMAXPROCS

	// set number of thread based on quotas or available threads in machine
	if _, err := maxprocs.Set(); err != nil {
		return fmt.Errorf("maxprocs: %w", err)
	}

	log.Infow("startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	// =========================================================================
	// Start debug service
	log.Infow("startup", "status", "debug router started", "host", cfg.Web.DebugHost)

	// The debug function returns a mux to listen and serve on for all the debug
	// related endpoints. This included the standard library endpoints.

	// Construct the mux for the debug calls.
	debugMux := handlers.DebugStandardLibraryMux()

	// Start the service listening for debug request.
	// Not concerned with shutting this down with load shedding.
	go func() {
		if err := http.ListenAndServe(cfg.Web.DebugHost, debugMux); err != nil {
			log.Errorw("shutdown", "status", "debug router closed", "host", cfg.Web.DebugHost, "ERROR", err)
		}
	}()

	// =========================================================================
	// App Starting

	// log.Infow("starting service", "version", build)
	// defer log.Infow("shutdown complete")

	// out, err := conf.String(&cfg)
	// if err != nil {
	// 	return fmt.Errorf("generating config for output: %w", err)
	// }
	// log.Infow("startup", "config", out)

	// expvar.NewString("build").Set(build)

	// =========================================================================

	// shutdown := make(chan os.Signal, 1)
	// signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// <-shutdown

	// =========================================================================
	// Start API Service

	log.Infow("startup", "status", "initializing API support")

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Construct a server to service the requests against the mux.
	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      nil,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
		ErrorLog:     zap.NewStdLog(log.Desugar()),
	}

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Start the service listening for api requests.
	go func() {
		log.Infow("startup", "status", "api router started", "host", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	// =========================================================================
	// Shutdown

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Infow("shutdown", "status", "shutdown started", "signal", sig)
		defer log.Infow("shutdown", "status", "shutdown complete", "signal", sig)

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		// Asking listener to shutdown and shed load.
		if err := api.Shutdown(ctx); err != nil {
			api.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}
