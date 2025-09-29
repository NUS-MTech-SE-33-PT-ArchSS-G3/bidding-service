package config

type Cors struct {
	IsEnabled        bool
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	AllowMaxAge      int
}
