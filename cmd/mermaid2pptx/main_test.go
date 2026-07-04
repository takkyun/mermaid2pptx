package main

import (
	"flag"
	"reflect"
	"testing"
)

// TestParseArgs locks down that options are honored no matter where they sit
// relative to the input files (before, after, or between). Regression guard
// for the interspersed-flag handling in parseArgs.
func TestParseArgs(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantInputs []string
		wantOut    string
		wantForce  bool
		wantMargin float64
	}{
		{
			name:       "flags before input",
			args:       []string{"-f", "-o", "out.pptx", "a.svg"},
			wantInputs: []string{"a.svg"},
			wantOut:    "out.pptx",
			wantForce:  true,
			wantMargin: 0.3,
		},
		{
			name:       "flags after input",
			args:       []string{"a.svg", "-f", "-o", "out.pptx"},
			wantInputs: []string{"a.svg"},
			wantOut:    "out.pptx",
			wantForce:  true,
			wantMargin: 0.3,
		},
		{
			name:       "flags between inputs",
			args:       []string{"a.svg", "-margin", "0.5", "b.mmd"},
			wantInputs: []string{"a.svg", "b.mmd"},
			wantOut:    "",
			wantForce:  false,
			wantMargin: 0.5,
		},
		{
			name:       "equals form after input",
			args:       []string{"a.svg", "-o=out.pptx", "-margin=0.25"},
			wantInputs: []string{"a.svg"},
			wantOut:    "out.pptx",
			wantForce:  false,
			wantMargin: 0.25,
		},
		{
			name:       "no flags",
			args:       []string{"a.svg", "b.svg"},
			wantInputs: []string{"a.svg", "b.svg"},
			wantMargin: 0.3,
		},
		{
			name:       "no args",
			args:       []string{},
			wantInputs: nil,
			wantMargin: 0.3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			cfg, inputs := parseArgs(fs, tt.args)
			if !reflect.DeepEqual(inputs, tt.wantInputs) {
				t.Errorf("inputs = %#v, want %#v", inputs, tt.wantInputs)
			}
			if cfg.out != tt.wantOut {
				t.Errorf("out = %q, want %q", cfg.out, tt.wantOut)
			}
			if cfg.force != tt.wantForce {
				t.Errorf("force = %v, want %v", cfg.force, tt.wantForce)
			}
			if cfg.margin != tt.wantMargin {
				t.Errorf("margin = %v, want %v", cfg.margin, tt.wantMargin)
			}
		})
	}
}
