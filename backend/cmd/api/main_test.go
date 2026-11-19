package main

import (
	"testing"
	"time"
)

func TestFortalezaTimezoneIsAvailable(t *testing.T) {
	location, err := time.LoadLocation("America/Fortaleza")
	if err != nil {
		t.Fatalf("LoadLocation() error = %v", err)
	}
	if location.String() != "America/Fortaleza" {
		t.Errorf("location = %q, want America/Fortaleza", location)
	}
}
