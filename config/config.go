package config

// Config contains all wuerfler config settings
type Config struct {
	Port           int    `default:"80"`
	SecurePort     int    `default:"0"`
	SecureHostname string `default:""`
	CertDir        string `default:""`
	FrontendDir    string `default:""`
	Debug          bool
}
