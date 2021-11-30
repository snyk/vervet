package lib

type ServerConfig struct {
	Host string `json:"host"`
	Services []string `json:"services"`
}
