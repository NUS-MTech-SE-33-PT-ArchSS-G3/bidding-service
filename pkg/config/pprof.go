package config

type Pprof struct {
	IsEnabled bool `json:"IsEnabled"`
	Port      int  `json:"Port"`
}
