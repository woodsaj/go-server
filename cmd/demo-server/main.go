package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	// self registering services
	_ "github.com/woodsaj/go-server/api"
	_ "github.com/woodsaj/go-server/processor/processorbar"
	_ "github.com/woodsaj/go-server/processor/processorfoo"
	_ "github.com/woodsaj/go-server/workers/workera"
	_ "github.com/woodsaj/go-server/workers/workerb"
)

func init() {
	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the info severity or above.
	log.SetLevel(log.InfoLevel)

}

func main() {
	var logLevel string
	var confDir string
	flag.StringVar(&logLevel, "log-level", "info", "One of debug,info,warn,error,fatal,panic")
	flag.StringVar(&confDir, "config-dir", "/etc/demo", "path to configuration dir")
	flag.Parse()

	lvl, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Fatal(err)
	}
	log.SetLevel(lvl)

	// initialize our config
	if _, err := os.Stat(confDir); err == nil {
		viper.SetConfigName("config")
		viper.AddConfigPath(confDir)
		err := viper.ReadInConfig()
		if err != nil {
			log.Fatal(err)
		}
	}
	viper.SetEnvPrefix("DEMO")
	viper.AutomaticEnv()
	replacer := strings.NewReplacer("-", "_", ".", "_")
	viper.SetEnvKeyReplacer(replacer)

	srv := NewCoreSrv()
	go listenToSystemSignals(srv)

	log.Exit(srv.Exit(srv.Run()))
}

func listenToSystemSignals(srv *CoreSrv) {
	signalChan := make(chan os.Signal, 1)
	ignoreChan := make(chan os.Signal, 1)

	signal.Notify(ignoreChan, syscall.SIGHUP)
	signal.Notify(signalChan, os.Interrupt, os.Kill, syscall.SIGTERM)

	select {
	case sig := <-signalChan:
		srv.Shutdown(fmt.Sprintf("System signal: %s", sig))
	}
}
