package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
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
		optConf string
		optDry  bool
	)
	flag.StringVar(&optConf, "conf", "/etc/dellog.d", "directory to load configuration from")
	flag.BoolVar(&optDry, "dry", false, "dry run")
	flag.Parse()

	rules := rg.Must(LoadRuleDir(optConf))

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
					if !optDry {
						if err := os.Remove(file); err != nil {
							log.Printf("failed to remove %s: %v", file, err)
						}
					}
				case ActionTruncate:
					log.Println("truncate:", file)
					if !optDry {
						if err := os.Truncate(file, 0); err != nil {
							log.Printf("failed to truncate %s: %v", file, err)
						}
					}
				}
			}
		}
	}
}
