package main

import (
	"path/filepath"
	"strconv"
	"strings"
)

var (
	capacitySuffixes = map[string]int64{
		"k": 1024,
		"m": 1024 * 1024,
		"g": 1024 * 1024 * 1024,
		"t": 1024 * 1024 * 1024 * 1024,
	}
)

// Capacity is a capacity type
type Capacity int64

func (c *Capacity) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v int64
	if err := unmarshal(&v); err != nil {
		var s string
		if err := unmarshal(&s); err != nil {
			return err
		}
		v, err := ParseCapacity(s)
		if err != nil {
			return err
		}
		*c = v
		return nil
	}
	*c = Capacity(v)
	return nil
}

func (c Capacity) MarshalYAML() (interface{}, error) {
	return int64(c), nil
}

func (c Capacity) Unwrap() int64 {
	return int64(c)
}

// ParseCapacity parses capacity string
func ParseCapacity(s string) (capacity Capacity, err error) {
	s = strings.ToLower(s)

	var factor int64
	for suffix, _factor := range capacitySuffixes {
		if strings.HasSuffix(s, suffix) {
			factor = _factor
			s = strings.TrimSuffix(s, suffix)
			break
		}
	}

	var v int64

	if v, err = strconv.ParseInt(s, 10, 64); err != nil {
		return
	}

	if factor > 0 {
		v *= factor
	}

	capacity = Capacity(v)
	return
}

type SingleOrMulti[T any] []T

func (sm *SingleOrMulti[T]) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var v T
	if err := unmarshal(&v); err != nil {
		var vs []T
		if err := unmarshal(&vs); err != nil {
			return err
		}
		*sm = vs
	} else {
		*sm = []T{v}
	}
	return nil
}

func (sm SingleOrMulti[T]) Unwrap() []T {
	out := make([]T, len(sm))
	copy(out, sm)
	return out
}

// ExpandDoubleAsteriskPattern expands '**' in glob patterns to multiple '*'
func ExpandDoubleAsteriskPattern(patterns *[]string) {
	var expanded []string
	for _, rule := range *patterns {
		if strings.Contains(rule, "**"+string(filepath.Separator)) {
			for i := 0; i < 5; i++ {
				expanded = append(expanded,
					strings.ReplaceAll(
						rule,
						"**"+string(filepath.Separator),
						strings.Repeat("*"+string(filepath.Separator), i),
					),
				)
			}
		} else {
			expanded = append(expanded, rule)
		}
	}
	*patterns = expanded
}
