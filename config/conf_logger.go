package config

type LoggerSetting struct {
	Level        string
	Prefix       string
	Director     string
	ShowLine     bool `mapstructure:"show_line"`
	LogInConsole bool `mapstructure:"log_in_console"`
}
