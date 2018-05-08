package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/leyantech/es-hit/es"
	"github.com/leyantech/es-hit/graphite"
	log "github.com/sirupsen/logrus"

	"flag"
)

type config struct {
	Graphite graphite.Config `toml:"graphite"`
	Rules    []es.Rule
}

var configFile string
var verbose bool

func init() {
	// log.SetFormatter(&log.JSONFormatter{})
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)

	flag.StringVar(&configFile, "config", "", "Configuration file")
	flag.BoolVar(&verbose, "verbose", false, "Verbose output")
	flag.Parse()
}

var conf config

func runSearch(ctx context.Context, wg *sync.WaitGroup, rule es.Rule, graphiteWorker *graphite.Worker) {
	defer wg.Done()

	// Create new ES Wrapper for this rule
	esWrapper, err := es.NewWrapper(&rule)
	if err != nil {
		log.Errorf("Failed to create ES Wrapper search, %v", err)
	}

	duration, err := time.ParseDuration(rule.CheckEvery)
	if err != nil {
		log.Fatalf("Can't not parse Check Interval string %s, %v", rule.CheckEvery, err)
	}
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			go func() {
				ctx, cancel := context.WithDeadline(ctx, time.Now().Add(duration))
				defer cancel()
				log.Debugf("Making search %s, against %s with pattern %s", rule.EsURL, rule.Index, rule.Pattern)
				hits, err := esWrapper.SearchHit(ctx)
				if err != nil {
					log.Errorf("Failed to do search, %v", err)
				}
				graphiteWorker.DoSend(rule.Name, float64(hits))
			}()
		case <-ctx.Done():
			log.Infof("ctx done %v", ctx.Err())
			return
		}
	}
}

func main() {

	if verbose {
		log.SetLevel(log.DebugLevel)
	}

	if configFile == "" {
		log.Fatalln("Please indicate config file with -config")
	}

	_, err := toml.DecodeFile(configFile, &conf)
	if err != nil {
		log.Fatalf("Failed to decode the TOML configuration file: %v, %v", configFile, err)
	}
	log.Debugln(conf)

	var wg sync.WaitGroup

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	graphiteWorker := graphite.NewWorker(conf.Graphite)
	ctx, cancel := context.WithCancel(context.Background())

	// Waiting for Program Inerrupt
	go func() {
		select {
		case <-signalCh:
			cancel()
		case <-ctx.Done():
			log.Infof("Parent ctx done %v", ctx.Err())
		}
	}()

	for _, rule := range conf.Rules {
		if !rule.Enabled {
			continue
		}
		wg.Add(1)
		log.Debugf("%v", rule)
		time.Sleep(time.Second)
		go runSearch(ctx, &wg, rule, graphiteWorker)
	}

	wg.Wait()
	log.Warnf("Program stop here")
}
