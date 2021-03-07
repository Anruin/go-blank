package monitoring

const (
	CfgHost    = "monitoring.host"
	CfgPort    = "monitoring.port"
	CfgTimeout = "monitoring.timeout"
)

// Monitoring configuration.
type Config struct {
	Host    string `mapstructure:"host"`
	Port    int    `mapstructure:"port"`
	Timeout int    `mapstructure:"timeout"`
}
