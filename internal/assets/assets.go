package assets

import "embed"

//go:embed config/*
var configFS embed.FS

func ReadConfigFile(name string) ([]byte, error) {
	return configFS.ReadFile("config/" + name)
}
