package utils_test

import (
	"testing"

	"github.com/AbePlays/go-url-shortener-system-design/utils"
)

func TestIsValidUrl(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "should return true for valid https url",
			url:  "https://www.wrongabhishek.com",
			want: true,
		},
		{
			name: "should return true for valid http url",
			url:  "http://example.com",
			want: true,
		},
		{
			name: "should return true for url with path and query",
			url:  "https://example.com/some/path?query=1",
			want: true,
		},
		{
			name: "should return false for missing scheme",
			url:  "www.wrongabhishek.com",
			want: false,
		},
		{
			name: "should return false for missing scheme bare domain",
			url:  "wrongabhishek.com",
			want: false,
		},
		{
			name: "should return false for unsupported ftp scheme",
			url:  "ftp://example.com",
			want: false,
		},
		{
			name: "should return false for javascript scheme",
			url:  "javascript:alert(1)",
			want: false,
		},
		{
			name: "should return false for empty host",
			url:  "https://",
			want: false,
		},
		{
			name: "should return false for empty string",
			url:  "",
			want: false,
		},
		{
			name: "should return false for garbage input with spaces",
			url:  "not a url at all with spaces",
			want: false,
		},
		{
			name: "should return false for some-random-string",
			url:  "some-random-string",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := utils.IsValidUrl(tt.url)
			if got != tt.want {
				t.Errorf("IsValidUrl() = %v, want %v", got, tt.want)
			}
		})
	}
}
