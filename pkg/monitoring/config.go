package monitoring

const (
	CfgHost = "monitoring.host"
	CfgPort = "monitoring.port"
)

// Monitoring configuration.
type Config struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}
