package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	stats := flag.Bool("stats", false, "display stats")
	ask := flag.String("ask", "", "ask a question")
	//use github.com/spf13/cobra if more feature are needed
	flag.Parse()

	if *ask != "" {
		response, err := getChatCompletionResponse(getPrompt(*ask))
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Println(response)
		}
	}

	if *stats {
		fmt.Println("Usage: gh-commit")
		numCommits, wordCount, err := getCommitStats()
		hoursSaved := calculateTimeSaved(numCommits, wordCount)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		} else {
			emoji := "🤖"
			message := fmt.Sprintf(
				"Git commit stats:\n"+
					"Number of commits: %d\n"+
					"Number of words in the commit message: %d\n"+
					"Format the git commit stats data and figure out profound insights on how writing these commit messages by AI is saving human hours? Use relevant emojis, use real-world stats in calculations and explain.\n"+
					"If all commit messages were written by %s, you would have saved %.1f hours! %s",
				numCommits, wordCount, "AI", hoursSaved, emoji)

			completionResponse, err := getChatCompletionResponse(getPrompt(message))
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			fmt.Println(completionResponse)
		}
	}

	if flag.NFlag() == 0 {
		diff, err := getGitDiff()
		if err != nil {
			// Check if it's just a truncation warning
			if _, ok := err.(*DiffTruncatedWarning); ok {
				// Print warning to stderr but continue
				fmt.Fprintf(os.Stderr, "%v\n", err)
			} else {
				// Real error - exit
				fmt.Printf("Error: %v\n", err)
				return
			}
		}
		completionResponse, err := getChatCompletionResponse(getDiffPrompt(diff))
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		completionResponse = formatResponse(completionResponse)
		fmt.Println(completionResponse)
	}
}
