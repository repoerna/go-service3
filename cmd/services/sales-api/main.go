package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

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
	// App Starting

	log.Infow("starting service", "version", build)
	defer log.Infow("shutdown complete")

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}
	log.Infow("startup", "config", out)

	// =========================================================================

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	<-shutdown

	// expvar.NewString("build").Set(build)

	return nil
}
