package utils

import "net/url"

func IsValidUrl(rawUrl string) bool {
	parsedUrl, err := url.ParseRequestURI(rawUrl)
	if err != nil {
		return false
	}

	if parsedUrl.Scheme != "http" && parsedUrl.Scheme != "https" {
		return false
	}

	if parsedUrl.Host == "" {
		return false
	}

	return true
}
