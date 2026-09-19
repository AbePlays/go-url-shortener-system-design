package utils_test

import (
	"testing"

	"github.com/AbePlays/go-url-shortener-system-design/utils"
)

func TestGenerateShortCode(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{
			name:   "should generate short code of specified length",
			length: 6,
		},
		{
			name:   "should generate short code of specified length",
			length: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.GenerateShortCode(tt.length)

			if len(got) != tt.length {
				t.Errorf("GenerateShortCode() = %v, want %v", got, tt.length)
			}
		})
	}
}
