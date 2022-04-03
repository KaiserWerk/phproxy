package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"

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

	bindAddr := prepareBindAddress(appConfig.PHProxy)

	base := handler.Base{
		Config: appConfig,
		Logger: logger,
	}

	router := mux.NewRouter()
	for _, app := range appConfig.Apps {
		logger.Printf("Host for app '%s': %s\n", app.Name, app.Host)
		remote, err := url.ParseRequestURI(app.Host)
		if err != nil {
			logger.Println("could not parse app host for app '%s': %s", app.Name, err.Error())
			continue
		}

		proxy, err := NewProxy(remote)
		if err != nil {
			logger.Println("could not create proxy for app '%s': %s", app.Name, err.Error())
			continue
		}

		logger.Println("remote Host:", remote.Host)
		origDir := proxy.Director
		proxy.Director = func(r *http.Request) { // vorerst
			origDir(r)
			r.URL.Host = remote.Host
			r.Host = remote.Host
			logger.Println("remote:", remote.String())
			if strings.HasSuffix(remote.Path, ".js") {
				logger.Printf("trying to GET javascript file %s", remote.String())
			}
		}
		proxy.ModifyResponse = func(resp *http.Response) error {
			//if resp.StatusCode >= 300 && resp.StatusCode < 400 { // a redirection
			//	l := resp.Header.Get("Location")
			//	if l != "" {
			//		u, _ := url.Parse(l)
			//		u.Scheme = bindAddr.Scheme
			//		u.Host = fmt.Sprintf("%s:%s", app.PublicDomain, bindAddr.Port())
			//		resp.Header.Set("Location", u.String())
			//		logger.Println("rewritten location header from", l, "to", u.String())
			//	}
			//}
			return nil
		}

		logger.Println("Public Domain:", app.PublicDomain)

		router.PathPrefix("/").HandlerFunc(base.Handler(proxy, app.Host)).Host(app.PublicDomain)

		/* TODO:
		bei response status >= 300 und < 400 mit ModifyResponse die response so ändern, dass
		schema, host und port richtig gesetzt sind!!
		*/

		//s := router.Host(app.PublicDomain).Subrouter()
		//s.HandleFunc("/", base.Handler(proxy, app.Host))
		// add more handle funcs

		// TODO: dont forget htaccess files

		//for from, to := range app.UrlRewriting.Custom {
		//	logger.Printf("[DEBUG] adding custom rule: '%s' to '%s'\n", from, to)
		//
		//	if strings.Index(from, "?") != -1 {
		//		from = strings.Split(from, "?")[0]
		//	}
		//
		//	var (
		//		parts   []string
		//		toPath  = to
		//		toQuery = ""
		//	)
		//	if strings.Index(to, "?") != -1 {
		//		parts = strings.Split(to, "?")
		//		toPath = parts[0]
		//		toQuery = parts[1]
		//	}
		//
		//
		//}
	}

	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		msg := fmt.Sprintf("Path '%s' is not registered for host '%s'", r.URL.Path, r.URL.Hostname())
		fmt.Fprint(w, msg)
		logger.Println("[DEBUG] " + msg)
	})

	srv := http.Server{
		Handler:      router,
		Addr:         bindAddr.Host,
		WriteTimeout: 3 * time.Minute,
		ReadTimeout:  2 * time.Minute,
	}

	exitCh := make(chan os.Signal, 1)
	signal.Notify(exitCh, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-exitCh
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()
		if srv.Shutdown(ctx) != nil {
			logger.Println("could not shut down server:", err.Error())
		} else {
			logger.Println("server shutdown complete")
		}
	}()

	if appConfig.PHProxy.TLS.CertFile != "" && appConfig.PHProxy.TLS.KeyFile != "" {
		logger.Printf("starting up HTTP server with TLS using bind address '%s'...\n", appConfig.PHProxy.BindAddr)
		err = srv.ListenAndServeTLS(appConfig.PHProxy.TLS.CertFile, appConfig.PHProxy.TLS.KeyFile)
	} else {
		logger.Printf("starting up HTTP server using bind address '%s'...\n", appConfig.PHProxy.BindAddr)
		err = srv.ListenAndServe()
	}
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Println("server error:", err.Error())
	}
}

func prepareBindAddress(config config.PHProxy) *url.URL {
	u := new(url.URL)
	u.Host = config.BindAddr
	if config.TLS.CertFile != "" && config.TLS.KeyFile != "" {
		u.Scheme = "https"
	} else {
		u.Scheme = "http"
	}
	return u
}

func NewProxy(appAddress *url.URL) (*httputil.ReverseProxy, error) {
	proxy := httputil.NewSingleHostReverseProxy(appAddress)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		fmt.Fprintf(w, "proxy error: %s", err.Error())
	}

	return proxy, nil
}

func director(originalDirector func(*http.Request), originalHost string, r *http.Request, toPath, toQuery string) {
	originalDirector(r)
	fmt.Printf("rewriting url from '%s'", r.URL.String())
	vars := mux.Vars(r)
	for name, value := range vars {
		toPath = strings.ReplaceAll(toPath, "{"+name+"}", value)
		toQuery = strings.ReplaceAll(toQuery, name, strings.Trim(value, "{}"))
	}
	r.URL.Path = toPath
	r.URL.RawQuery = toQuery
	fmt.Printf(" to '%s'\n", r.URL.String())
	r.Host = originalHost // wichtig, da sonst für lokale Domains nur die Loopback-Adresse verwendet wird!
	fmt.Println("\n host:", r.Host)

}
