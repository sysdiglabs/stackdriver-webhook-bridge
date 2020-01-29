package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/sysdiglabs/stackdriver-webhook-bridge/model"
	"github.com/sysdiglabs/stackdriver-webhook-bridge/poller"

	log "github.com/sirupsen/logrus"
)

func main() {

	var err error

	cfg := model.NewConfig()

	dur, err := time.ParseDuration("5s")
	if err != nil {
		log.Fatalf("Could not parse duration '5s': %v", err)
	}

	flag.StringVar(&cfg.Url, "url", "http://sysdig-agent.sysdig-agent.svc.cluster.local:7765/k8s_audit", "send generated events to this url")
	flag.StringVar(&cfg.ProjectArg, "project", "", "read logs from provided project. If blank, use metadata service to find project id")
	flag.StringVar(&cfg.LogfileName, "logfile", "", "if set, write all log entries to provided file")
	flag.StringVar(&cfg.OutfileName, "outfile", "", "if set, also append converted audit logs to provided file")
	flag.DurationVar(&cfg.PollInterval, "poll_interval", dur, "poll interval for log messages")
	flag.StringVar(&cfg.LogLevel, "log_level", "info", "log level")

	log.SetLevel(log.InfoLevel)
	flag.Parse()

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

	curTime := time.Now().UTC().Add(-1 * cfg.PollInterval)

	loopChan := make(chan string)
	signalChan := make(chan os.Signal, 2)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChan
		loopChan <- "exit"
	}()

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
