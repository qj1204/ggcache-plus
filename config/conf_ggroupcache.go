package config

type GGroupCache struct {
	Name            string
	Addr            []string
	TTL             int
	CleanUpInterval int `mapstructure:"clean_up_interval"`
}
