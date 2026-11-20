package main

import (
	"path/filepath"
	"testing"
)

func TestParseOptions(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want options
	}{
		{
			name: "padrões",
			args: nil,
			want: options{file: "../design-system/seed.json"},
		},
		{
			name: "reset e dry-run",
			args: []string{"-reset", "-dry-run"},
			want: options{file: "../design-system/seed.json", reset: true, dryRun: true},
		},
		{
			name: "arquivo customizado",
			args: []string{"-file", "/tmp/seed.json"},
			want: options{file: "/tmp/seed.json"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseOptions(test.args)
			if err != nil {
				t.Fatalf("parseOptions() error = %v", err)
			}
			if got != test.want {
				t.Fatalf("parseOptions() = %+v, want %+v", got, test.want)
			}
		})
	}
}

func TestParseOptionsRejectsUnknownFlag(t *testing.T) {
	if _, err := parseOptions([]string{"-unknown"}); err == nil {
		t.Fatal("parseOptions() error = nil, want error")
	}
}

func TestRunDryRunValidatesRepositorySeed(t *testing.T) {
	path := filepath.Join("..", "..", "..", "design-system", "seed.json")

	if err := run([]string{"-file", path, "-dry-run"}, nil); err != nil {
		t.Fatalf("run() error = %v", err)
	}
}
