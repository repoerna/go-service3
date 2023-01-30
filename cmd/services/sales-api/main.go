package main

import (
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"go.uber.org/automaxprocs/maxprocs"
)

var build = "develop"

func main() {
	_, err := maxprocs.Set()
	if err != nil {
		log.Printf("maxprocs: %w", err)
		os.Exit(1)
	}

	g := runtime.GOMAXPROCS(0)

	log.Printf("starting service build[%s] CPU[%d]", build, g)
	defer log.Println("service ended")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	<-shutdown

	log.Println("stopping service")

	// log, err := logger.New("SALES_API")
	// if err != nil {
	// 	fmt.Println("error creating logger: ", err)
	// }

	// log.Info("test")
}
