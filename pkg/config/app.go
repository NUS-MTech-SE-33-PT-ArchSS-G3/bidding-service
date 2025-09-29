package config

type App struct {
	Name        string
	Environment Environment
	Version     string
}

type Environment string

const (
	Dev  Environment = "development"
	Prod Environment = "production"
)
