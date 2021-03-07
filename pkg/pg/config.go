package pg

const (
	CfgHost = "pg.host"
	CfgPort = "pg.port"
	CfgName = "pg.name"
	CfgUser = "pg.user"
	CfgPass = "pg.pass"
)

type Config struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Name string `mapstructure:"name"`
	User string `mapstructure:"user"`
	Pass string `mapstructure:"pass"`
}
