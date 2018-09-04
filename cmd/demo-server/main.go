package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	ini "github.com/glacjay/goini"
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
	var oldConfig bool
	flag.StringVar(&logLevel, "log-level", "info", "One of debug,info,warn,error,fatal,panic")
	flag.StringVar(&confDir, "config-dir", "/etc/demo", "path to configuration dir")
	flag.BoolVar(&oldConfig, "old-config", true, "use old .ini config file")
	flag.Parse()

	lvl, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Fatal(err)
	}
	log.SetLevel(lvl)

	// initialize our config
	if !oldConfig {
		viper.SetConfigName("config")
		viper.AddConfigPath(confDir)
		err = viper.ReadInConfig()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		// load old ini config
		iniConf, err := ini.Load(confDir + "/config.ini")
		if err != nil {
			log.Fatal(err)
		}
		// now write the ini file out to a []byte in YAML format.
		yaml := new(bytes.Buffer)
		for name, section := range iniConf {
			for key, val := range section {
				if name == "" {
					yaml.WriteString(fmt.Sprintf("%s: %s\n", key, val))
				} else {
					yaml.WriteString(fmt.Sprintf("%s.%s: %s\n", name, key, val))
				}
			}

		}
		fmt.Println(yaml.String())
		viper.SetConfigType("yaml")
		viper.ReadConfig(yaml)
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
