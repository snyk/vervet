package lib

type ServerConfig struct {
	Services []string `json:"services"`
}

type VersionList struct {
	Versions []string
}