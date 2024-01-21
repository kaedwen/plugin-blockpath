// Package plugin_blockpath a plugin to block a path.
package plugin_blockpath

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
)

// Config holds the plugin configuration.
type Config struct {
	Allows []string `json:"allows,omitempty"`
	Blocks []string `json:"blocks,omitempty"`
}

// CreateConfig creates and initializes the plugin configuration.
func CreateConfig() *Config {
	return &Config{}
}

type blockPath struct {
	name   string
	next   http.Handler
	allows []*regexp.Regexp
	blocks []*regexp.Regexp
}

func prepare(l []string) ([]*regexp.Regexp, error) {
	regexps := make([]*regexp.Regexp, len(l))

	for i, regex := range l {
		re, err := regexp.Compile(regex)
		if err != nil {
			return nil, fmt.Errorf("error compiling regex %q: %w", regex, err)
		}

		regexps[i] = re
	}

	return regexps, nil
}

// New creates and returns a plugin instance.
func New(_ context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	allows, err := prepare(config.Allows)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare allow regex - %w", err)
	}

	blocks, err := prepare(config.Blocks)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare allow regex - %w", err)
	}

	return &blockPath{name, next, allows, blocks}, nil
}

func (b *blockPath) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	currentPath := req.URL.EscapedPath()

	for _, re := range b.allows {
		if re.MatchString(currentPath) {
			b.next.ServeHTTP(rw, req)
			return
		}
	}

	for _, re := range b.blocks {
		if re.MatchString(currentPath) {
			rw.WriteHeader(http.StatusForbidden)
			return
		}
	}

	b.next.ServeHTTP(rw, req)
}
