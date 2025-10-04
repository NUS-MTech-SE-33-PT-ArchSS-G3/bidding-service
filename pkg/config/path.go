package config

import (
	"flag"
	"os"
	"path/filepath"
)

// Path resolves the config file location
func Path() string {
	// 1st pref, explicit env
	if p := os.Getenv("CONFIG_PATH"); p != "" {
		return p
	}

	// 2nd pref: -config flag value
	if flag.Parsed() {
		if f := flag.Lookup("config"); f != nil {
			if v := f.Value.String(); v != "" {
				return v
			}
		}
	}

	// 3rd pref: container default
	if _, err := os.Stat("/app/config.json"); err == nil {
		return "/app/config.json"
	}

	// 4th pref: next to exe
	if exe, err := os.Executable(); err == nil {
		return filepath.Join(filepath.Dir(exe), "config.json")
	}
	return "config.json"
}
