package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestNetworkStruct(t *testing.T) {
	network := Network{
		Port:        8080,
		IsLocalHost: true,
		Ssl: Ssl{
			IsEnabled: true,
			CertFile:  "/path/to/cert.pem",
			KeyFile:   "/path/to/key.pem",
			CAFile:    "/path/to/ca.pem",
		},
	}

	if network.Port != 8080 {
		t.Errorf("expected Port to be 8080, got %d", network.Port)
	}

	if !network.IsLocalHost {
		t.Error("expected IsLocalHost to be true")
	}

	if !network.Ssl.IsEnabled {
		t.Error("expected Ssl.IsEnabled to be true")
	}

	if network.Ssl.CertFile != "/path/to/cert.pem" {
		t.Errorf("expected CertFile to be '/path/to/cert.pem', got %s", network.Ssl.CertFile)
	}
}

func TestBindSsl(t *testing.T) {
	v := viper.New()
	
	// Call BindSsl
	BindSsl(v)

	// Set environment variables to test binding
	t.Setenv("SSL_CERT_FILE", "/test/cert.pem")
	t.Setenv("SSL_KEY_FILE", "/test/key.pem")
	t.Setenv("SSL_CA_FILE", "/test/ca.pem")

	v.AutomaticEnv()

	// Test that environment variables are bound correctly
	certFile := v.GetString("network.ssl.cert_file")
	keyFile := v.GetString("network.ssl.key_file")
	caFile := v.GetString("network.ssl.ca_file")

	if certFile != "/test/cert.pem" {
		t.Errorf("expected cert_file to be '/test/cert.pem', got %s", certFile)
	}

	if keyFile != "/test/key.pem" {
		t.Errorf("expected key_file to be '/test/key.pem', got %s", keyFile)
	}

	if caFile != "/test/ca.pem" {
		t.Errorf("expected ca_file to be '/test/ca.pem', got %s", caFile)
	}
}
