package main

import (
	"bytes"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type RuleDoc struct {
	Match SingleOrMulti[string] `yaml:"match,omitempty"`
	Days  int64                 `yaml:"days,omitempty"`
	Size  Capacity              `yaml:"size,omitempty"`
}

func ParseRule(buf []byte) (rules []RuleDoc, err error) {
	dec := yaml.NewDecoder(bytes.NewReader(buf))
	for {
		var rule RuleDoc
		if err = dec.Decode(&rule); err != nil {
			if err == io.EOF {
				err = nil
			}
			return
		}
		if len(rule.Match) == 0 {
			continue
		}
		rules = append(rules, rule)
	}
}

type Rule struct {
	Match []string
	Days  int64
	Size  int64
}

func LoadRuleDir(dir string) (rules []Rule, err error) {
	var entries []os.DirEntry

	if entries, err = os.ReadDir(dir); err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		log.Println("loading rules:", entry.Name())

		var buf []byte
		if buf, err = os.ReadFile(filepath.Join(dir, entry.Name())); err != nil {
			return
		}

		var fileRules []RuleDoc
		if fileRules, err = ParseRule(buf); err != nil {
			return
		}

		for _, rule := range fileRules {
			if len(rule.Match) == 0 || (rule.Size <= 0 && rule.Days <= 0) {
				log.Println("invalid rule:", rule)
				continue
			}
			patterns := rule.Match.Unwrap()
			ExpandDoubleAsteriskPattern(&patterns)
			rules = append(rules, Rule{
				Match: patterns,
				Days:  rule.Days,
				Size:  rule.Size.Unwrap(),
			})
		}
	}
	return
}

type Action int

const (
	ActionSkip Action = iota
	ActionRemove
	ActionTruncate
)

func EvaluateRule(name string, rule Rule, now time.Time) Action {
	if rule.Days > 0 {
		if date, err := DateFromFilename(name); err == nil {
			if now.Sub(date) > time.Duration(rule.Days)*24*time.Hour {
				return ActionRemove
			}
		}
	}
	if rule.Size > 0 {
		if info, err := os.Stat(name); err == nil {
			if info.Size() > rule.Size {
				return ActionTruncate
			}
		}
	}
	return ActionSkip
}
