package utils

import "strings"

func ExtractTextLines(text string) []string {
	var extractedLines []string
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "" {
			extractedLines = append(extractedLines, trimmedLine)
		}
	}

	return extractedLines
}
