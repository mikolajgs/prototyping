package prototyping

type DbConfig struct {
	Host string
	Port string
	User string
	Pass string
	Name string
	TablePrefix string
}

type HttpConfig struct {
	Port string
	ApiUri string
	UiUri string
}
