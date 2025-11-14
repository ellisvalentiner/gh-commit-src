package main

import (
	"flag"
	"testing"
)

func TestStatsFlag(t *testing.T) {
	flagSet := flag.NewFlagSet("TestStatsFlag", flag.ContinueOnError)
	stats := flagSet.Bool("stats", false, "get stats")

	err := flagSet.Parse([]string{"-stats"})
	if err != nil {
		t.Fatal("Error parsing flags:", err)
	}

	if *stats != true {
		t.Error("Expected stats flag to be true, but it was false")
	}
}

func TestAskFlag(t *testing.T) {
	flagSet := flag.NewFlagSet("TestAskFlag", flag.ContinueOnError)
	ask := flagSet.String("ask", "", "ask a question")

	err := flagSet.Parse([]string{"-ask", "test question"})
	if err != nil {
		t.Fatal("Error parsing flags:", err)
	}

	if *ask != "test question" {
		t.Errorf("Expected ask flag to be 'test question', but got '%s'", *ask)
	}
}

// Note: Testing main() directly is difficult as it has side effects
// In a production environment, you'd refactor main() to be more testable
// by extracting the logic into separate functions
