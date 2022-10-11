package database

type Config struct {
	Mongo struct {
		UserName string
		Password string
		Hostname string
		Port     int
		Database string
	}
	Redis struct {
		Password string
		Hostname string
		Port     int
	}
	Baidu struct {
		Ak string
		Sk string
	}
}

var DefaultConfig Config
