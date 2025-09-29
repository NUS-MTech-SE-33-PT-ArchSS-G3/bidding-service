package config

import "github.com/spf13/viper"

type Network struct {
	Port        int
	IsLocalHost bool
	Ssl         Ssl
}

type Ssl struct {
	IsEnabled bool
	CertFile  string `json:"-" mapstructure:"cert_file" env:"SSL_CERT_FILE"`
	KeyFile   string `json:"-" mapstructure:"key_file" env:"SSL_KEY_FILE"`
	CAFile    string `json:"-" mapstructure:"ca_file" env:"SSL_CA_FILE"`
}

func BindSsl(v *viper.Viper) {
	_ = v.BindEnv("network.ssl.cert_file", "SSL_CERT_FILE")
	_ = v.BindEnv("network.ssl.key_file", "SSL_KEY_FILE")
	_ = v.BindEnv("network.ssl.ca_file", "SSL_CA_FILE")
}
