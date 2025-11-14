package main

import (
	"io/ioutil"
	"log"
	"strings"
)

// GetAgentVersion returns the agent version from VERSION file
// The VERSION file is created during installation with the downloaded version
func GetAgentVersion() string {
	// VERSION file is created by installer at /Applications/ROI Agent/Resources/VERSION
	versionPath := "/Applications/ROI Agent/Resources/VERSION"

	if content, err := ioutil.ReadFile(versionPath); err == nil {
		version := strings.TrimSpace(string(content))
		if version != "" {
			log.Printf("Agent version: %s (from %s)", version, versionPath)
			return version
		}
	}

	// Fallback for development environment
	log.Printf("Agent version: dev (VERSION file not found, likely running in development)")
	return "dev"
}
