package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/sysdiglabs/stackdriver-webhook-bridge/config"
	"github.com/sysdiglabs/stackdriver-webhook-bridge/poller"
	"github.com/sysdiglabs/stackdriver-webhook-bridge/prometheus"

	pflag "github.com/spf13/pflag"
	log "github.com/sirupsen/logrus"
)

func main() {

	var err error

	log.SetLevel(log.InfoLevel)

	pflag.String("config", "/opt/swb/config/", "Read config file 'swb-config.yaml' below this directory")
	pflag.String("url", "http://sysdig-agent.sysdig-agent.svc.cluster.local:7765/k8s_audit", "send generated events to this url")
	pflag.String("project", "", "read logs from provided project. If blank, use metadata service to find project id")
	pflag.String("cluster", "", "read logs for provided cluster name. If blank, use metadata service to find cluster name")
	pflag.String("logfile", "", "if set, write all log entries to provided file")
	pflag.String("outfile", "", "if set, also append converted audit logs to provided file")
	pflag.Duration("poll_interval", 5 * time.Second, "poll interval for log messages")
	pflag.Duration("lag_interval", 30 * time.Second, "lag behind current time when reading log entries")
	pflag.String("log_level", "info", "log level")

	pflag.Parse()

	configPath, err := pflag.CommandLine.GetString("config")
	if err != nil {
		log.Fatalf("Could not get config path: %v", err)
	}

	cfg, err := config.New(configPath, pflag.CommandLine)

	if err != nil {
		log.Fatalf("Could not read configuration: %v", err)
	}

	level, err := log.ParseLevel(strings.ToUpper(cfg.LogLevel))
	if err != nil {
		log.Fatalf("Could not parse log level: %v", err)
	}

	log.SetLevel(level)

	ctx := context.Background()

	log.Debugf("Creating poller...")

	pollr, err := poller.NewPoller(ctx, cfg)

	if err != nil {
		log.Fatalf("Could not create poller: %v", err)
	}

	defer pollr.Close()

	curTime := time.Now().UTC().Add(-2 * cfg.LagInterval)

	loopChan := make(chan string)
	signalChan := make(chan os.Signal, 2)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		loopChan <- "exit"
	}()

	go prometheus.ExposeMetricsEndpoint(cfg.PrometheusPort)
	for {
		curTime = pollr.PollLogsSendEvents(curTime)
		go func() {
			time.Sleep(cfg.PollInterval)
			loopChan <- "timeout"
		}()

		msg := <-loopChan

		if msg == "exit" {
			break
		}
	}

	log.Infof("Done.")

	os.Exit(0)
}
