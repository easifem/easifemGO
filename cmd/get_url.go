package cmd

import (
	"errors"

	"github.com/spf13/viper"
)

// this function returns the url from config file
func get_url(a, b string) (string, error) {
	opts := []string{
		"git",
		"url",
	}

	for _, opt := range opts {
		key := a + "." + b + "." + opt
		if viper.IsSet(key) {
			return viper.GetString(key), nil
		}

	}
	return "", errors.New("no url related tag found")
}
