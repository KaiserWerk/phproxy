package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KaiserWerk/phproxy/internal/assets"
	"github.com/KaiserWerk/phproxy/internal/config"
	"github.com/KaiserWerk/phproxy/internal/handler"
	"github.com/KaiserWerk/phproxy/internal/logging"
)

var (
	createConfig = flag.Bool("create-config", false, "Wheather you want to create an example config file")
	configFile   = flag.String("cfg", "./phproxy.yaml", "The configuration file to use")
	logDir       = flag.String("logs", ".", "The directory to write log files into")
)

func main() {
	flag.Parse()

	if *createConfig {
		dist, _ := assets.ReadConfigFile("phproxy.dist.yaml")
		if err := os.WriteFile(*configFile, dist, 0644); err != nil {
			fmt.Println("could not create configuration file:", err.Error())
			os.Exit(-1)
		}

		fmt.Printf("configuration file '%s' created", *configFile)
		os.Exit(0)
	}

	appConfig, err := config.Init(*configFile)
	if err != nil {
		fmt.Println("could not create logger:", err.Error())
		os.Exit(-1)
	}

	if len(appConfig.Apps) == 0 {
		fmt.Println("could not find any PHP apps defined in configuration")
		os.Exit(-1)
	}

	var (
		logger  *log.Logger
		cleanup func() error
	)
	logger, cleanup, err = logging.New(*logDir)
	if err != nil {
		logger = log.New(os.Stdout, "", 0)
		logger.Println("using fallback logger")
	}
	defer func() {
		if cleanup != nil {
			if err := cleanup(); err != nil {
				fmt.Println("cleanup error:", err.Error())
			}
		}
	}()

	base := handler.Base{
		Config: appConfig,
		Logger: logger,
	}

	srv := http.Server{
		Handler:      http.HandlerFunc(base.Handler),
		Addr:         appConfig.PHProxy.BindAddr,
		WriteTimeout: 3 * time.Minute,
		ReadTimeout:  2 * time.Minute,
	}

	exitCh := make(chan os.Signal, 1)
	signal.Notify(exitCh, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-exitCh
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			logger.Println("could not shut down server:", err.Error())
		} else {
			logger.Println("server shutdown complete")
		}
	}()

	if appConfig.PHProxy.TLS.CertFile != "" && appConfig.PHProxy.TLS.KeyFile != "" {
		logger.Printf("starting up using bind address '%s'...\n", appConfig.PHProxy.BindAddr)
		err = srv.ListenAndServeTLS(appConfig.PHProxy.TLS.CertFile, appConfig.PHProxy.TLS.KeyFile)
	} else {
		err = srv.ListenAndServe()
	}
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Println("server error:", err.Error())
	}
}
