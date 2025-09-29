package config

type Logger struct {
	Level       string
	Format      string
	Output      string
	FilePath    string
	Environment Environment
}
