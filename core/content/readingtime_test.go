package content

import "testing"

func TestCountWords(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"plain text", "hello world foo bar", 4},
		{"with HTML tags", "<p>hello <strong>world</strong></p>", 2},
		{"empty", "", 0},
		{"only tags", "<div><span></span></div>", 0},
		{"nested tags", "<p>one <a href='#'>two <em>three</em></a> four</p>", 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CountWords(tt.input); got != tt.want {
				t.Errorf("CountWords() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCountImages(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"no images", "<p>text</p>", 0},
		{"one image", `<p><img src="a.png"></p>`, 1},
		{"multiple", `<img src="a.png"><p>text</p><img src="b.png">`, 2},
		{"self-closing", `<img src="a.png" />`, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CountImages(tt.input); got != tt.want {
				t.Errorf("CountImages() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCalculateReadingTime(t *testing.T) {
	tests := []struct {
		name   string
		words  int
		images int
		want   int
	}{
		{"minimum 1 min", 0, 0, 1},
		{"short text", 50, 0, 1},
		{"150 words = 1 min", 150, 0, 1},
		{"300 words = 2 min", 300, 0, 2},
		{"with images", 150, 5, 2},
		{"long article", 1500, 3, 11},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateReadingTime(tt.words, tt.images); got != tt.want {
				t.Errorf("CalculateReadingTime(%d, %d) = %d, want %d", tt.words, tt.images, got, tt.want)
			}
		})
	}
}
