package cache

import (
	"fmt"
	"strings"
	"testing"
)

func TestNodeCacheList(t *testing.T) {
	var cfg Config
	cfg = ForNode(cfg)

	var ww []string
	for _, w := range cfg.Watches {
		ww = append(ww, string(w.Kind))
	}

	fmt.Println("--->", strings.Join(ww, ","))
}
