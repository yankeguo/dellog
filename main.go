package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/yankeguo/rg"
)

func main() {
	var err error
	defer func() {
		if err == nil {
			return
		}
		log.Println("exited with error:", err.Error())
		os.Exit(1)
	}()
	defer rg.Guard(&err)

	var (
		optConf     string
		optDry      bool
		optInterval time.Duration
	)
	flag.StringVar(&optConf, "conf", "/etc/dellog.d", "directory to load configuration from")
	flag.BoolVar(&optDry, "dry", false, "dry run")
	flag.DurationVar(&optInterval, "interval", 0, "interval to rerun (e.g., 1h, 30m, 5s). If not set, run once and exit")
	flag.Parse()

	if optInterval > 0 {
		// Handle graceful shutdown on interrupt
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		go func() {
			<-sigChan
			log.Println("received interrupt signal, shutting down...")
			cancel()
		}()

		log.Printf("running in loop mode with interval: %v", optInterval)
		ticker := time.NewTicker(optInterval)
		defer ticker.Stop()

		// Run once immediately
		runOnce(optConf, optDry)

		// Then run on interval
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				runOnce(optConf, optDry)
			}
		}
	} else {
		runOnce(optConf, optDry)
	}
}

func runOnce(confDir string, dryRun bool) {
	rules := rg.Must(LoadRuleDir(confDir))
	now := time.Now()

	for _, rule := range rules {
		for _, match := range rule.Match {
			for _, file := range rg.Must(filepath.Glob(match)) {
				log.Println("found:", file)
				switch EvaluateRule(file, rule, now) {
				case ActionSkip:
					log.Println("skip:", file)
				case ActionRemove:
					log.Println("remove:", file)
					if !dryRun {
						if err := os.Remove(file); err != nil {
							log.Printf("failed to remove %s: %v", file, err)
						}
					}
				case ActionTruncate:
					log.Println("truncate:", file)
					if !dryRun {
						if err := os.Truncate(file, 0); err != nil {
							log.Printf("failed to truncate %s: %v", file, err)
						}
					}
				}
			}
		}
	}
}
