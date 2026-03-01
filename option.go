package gojand

import "github.com/microcosm-cc/bluemonday"

// Option configures a Transformer.
type Option func(*config)

type config struct {
	funcName   string
	globals    map[string]any
	policies   map[string]*bluemonday.Policy
	typescript bool
}

func newConfig(opts []Option) *config {
	cfg := &config{
		funcName: "transform",
		globals:  map[string]any{},
		policies: map[string]*bluemonday.Policy{},
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

// WithGlobal adds a global variable available to the script.
func WithGlobal(name string, value any) Option {
	return func(c *config) {
		c.globals[name] = value
	}
}

// WithGlobals adds multiple global variables available to the script.
func WithGlobals(globals map[string]any) Option {
	return func(c *config) {
		for k, v := range globals {
			c.globals[k] = v
		}
	}
}

// WithFuncName sets the name of the function to call in the script.
// Defaults to "transform".
func WithFuncName(name string) Option {
	return func(c *config) {
		c.funcName = name
	}
}

// WithPolicy registers a named bluemonday sanitization policy for use with
// html.strip_policy in scripts.
func WithPolicy(name string, policy *bluemonday.Policy) Option {
	return func(c *config) {
		c.policies[name] = policy
	}
}

// WithTypeScript enables esbuild TypeScript-to-JavaScript transpilation before
// script compilation.
func WithTypeScript() Option {
	return func(c *config) {
		c.typescript = true
	}
}
