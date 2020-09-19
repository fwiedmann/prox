package root

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/fwiedmann/prox/internal/infra"

	"github.com/fwiedmann/prox/internal/config"

	log "github.com/sirupsen/logrus"

	"github.com/fwiedmann/prox/internal/cache"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/fwiedmann/prox/domain/entity/route"
	"github.com/fwiedmann/prox/domain/usecase/configure"
	"github.com/fwiedmann/prox/domain/usecase/proxy"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.PersistentFlags().StringVar(&staticConfigFile, "static-config", "static.yaml", "Path to static config file")
	rootCmd.PersistentFlags().StringVar(&routesConfigFile, "routes-config", "routes.yaml", "Path to routes config file")
	rootCmd.PersistentFlags().StringVar(&tlsConfigFile, "tls-config", "tls.yaml", "Path to routes tls file")
	rootCmd.Flags().String("loglevel", "info", "Set a log level")

}

var staticConfigFile string
var routesConfigFile string
var tlsConfigFile string

var rootCmd = cobra.Command{
	Use:          "prox",
	Short:        "",
	SilenceUsage: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		cmdLogLevel, err := cmd.Flags().GetString("loglevel")
		if err != nil {
			return err
		}
		parsedLevel, err := log.ParseLevel(cmdLogLevel)
		if err != nil {
			return err
		}

		log.SetLevel(parsedLevel)
		log.SetFormatter(&log.TextFormatter{ForceColors: true})
		log.SetFormatter(&nested.Formatter{
			HideKeys: true,
		})
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		staticConfig, err := config.ParseStaticFile(staticConfigFile)
		if err != nil {
			return err
		}

		manager := route.NewManager(route.NewInMemRepo(), route.CreateHTTPClientForRoute)
		c := configure.NewFileConfigureUseCase(routesConfigFile, manager)

		configErr := make(chan error, 2)
		ctx, cancel := context.WithCancel(context.Background())
		go c.StartConfigure(ctx, configErr)

		tlsConf := config.NewDynamicTLSConfig(tlsConfigFile)

		go tlsConf.StartWatch(ctx, configErr)

		proxyErrorChan := make(chan error, len(staticConfig.Ports))

		for _, port := range staticConfig.Ports {
			go func(p config.Port) {
				px, err := proxy.NewUseCase(manager, configureCache(staticConfig.Cache.Enabled, staticConfig.Cache.CacheMaxSizeInMegaByte), p.Addr, staticConfig.AccessLogEnabled)
				if err != nil {
					proxyErrorChan <- err
					return
				}

				s := http.Server{
					Addr:    fmt.Sprintf(":%d", p.Addr),
					Handler: px,
				}

				if p.TlSEnabled {
					s.TLSConfig = &tls.Config{GetCertificate: tlsConf.GetCertificate}
					log.Debugf("Starting https endpoint on port %d", p.Addr)
					proxyErrorChan <- s.ListenAndServeTLS("", "")
					return
				}
				log.Debugf("Starting http endpoint on port %d", p.Addr)
				proxyErrorChan <- s.ListenAndServe()
			}(port)
		}

		go func() {
			proxyErrorChan <- infra.StartInfraHTTPEndpoint(int(staticConfig.InfraPort))
		}()

		osNotifyChan := initOSNotifyChan()

		select {
		case err := <-configErr:
			cancel()
			return err
		case err := <-proxyErrorChan:
			cancel()
			return err
		case osSignal := <-osNotifyChan:
			log.Warnf("received os %s signal, start  graceful shutdown of prox...", osSignal.String())
			cancel()
			return nil
		}
	},
}

func initOSNotifyChan() <-chan os.Signal {
	notifyChan := make(chan os.Signal, 3)
	signal.Notify(notifyChan, syscall.SIGTERM, syscall.SIGINT)
	return notifyChan
}

// Execute executes the rootCmd
func Execute() error {
	return rootCmd.Execute()
}

func configureCache(enabled bool, maxSize int64) proxy.Cache {
	if enabled {
		return cache.NewHTTPInMemoryCache(maxSize)
	}
	return cache.Empty{}
}
