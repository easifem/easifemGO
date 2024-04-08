package cmd

import (
	"errors"

	"github.com/spf13/viper"
)

// this function returns the url from config file
func install_get_url(pkg string) (string, error) {
	opts := []string{
		"git",
		"url",
	}

	for _, opt := range opts {
		if key := pkg + "." + opt; viper.IsSet(key) {
			return viper.GetString(key), nil
		}
		if key := "install." + pkg + "." + opt; viper.IsSet(key) {
			return viper.GetString(key), nil
		}
	}
	return "", errors.New("no url related tag found")
}
