package prototyping

type DBConfig struct {
	Host string
	Port string
	User string
	Pass string
	Name string
	TablePrefix string
}

type APIConfig struct {
	Port string
	URI string
}
