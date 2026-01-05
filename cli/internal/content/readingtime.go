package content

import (
	"math"
	"regexp"
	"strings"
)

const (
	// WordsPerMinute is the reading speed for dense/technical content
	// Lower than casual reading (200-250) to account for re-reading,
	// processing technical concepts, and following wiki-links
	WordsPerMinute = 150
	// SecondsPerImage is time spent viewing informative images (12 seconds)
	SecondsPerImage = 12
)

var imgTagRegex = regexp.MustCompile(`<img\s`)

// CountWords counts words in HTML content by stripping tags first
func CountWords(htmlContent string) int {
	// Strip HTML tags
	plain := htmlTagRegex.ReplaceAllString(htmlContent, " ")
	plain = htmlTagRegex.ReplaceAllString(plain, " ")

	// Use the same regex as page.go for consistency
	plain = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(htmlContent, " ")

	// Normalize whitespace and split
	words := strings.Fields(plain)
	return len(words)
}

// CountImages counts the number of <img> tags in HTML content
func CountImages(htmlContent string) int {
	return len(imgTagRegex.FindAllString(htmlContent, -1))
}

// CalculateReadingTime returns estimated reading time in minutes
// Formula: ceil((words / 200) + (images × 0.2))
// Minimum return value is 1 minute
func CalculateReadingTime(wordCount, imageCount int) int {
	// Time for words (in minutes)
	wordMinutes := float64(wordCount) / float64(WordsPerMinute)

	// Time for images (12 seconds = 0.2 minutes per image)
	imageMinutes := float64(imageCount) * float64(SecondsPerImage) / 60.0

	// Total time, rounded up
	total := int(math.Ceil(wordMinutes + imageMinutes))

	// Minimum 1 minute
	if total < 1 {
		return 1
	}
	return total
}
