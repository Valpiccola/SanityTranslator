package main

import "regexp"

// cleanString removes all non-alphabetic characters from the input string.
func cleanString(text string) string {
	return regexp.MustCompile(`[^a-zA-Z]+`).ReplaceAllString(text, "")
}
