package cmd

import (
	"log"

	"github.com/spf13/viper"
)

func get_string_slice_value(a, b, c string) []string {
	key := a + "." + b + "." + c

	if viper.IsSet(key) {
		return viper.GetStringSlice(key)
	}

	log.Println("[log] :: viper.GetString() cannot find ", key)
	return nil
}
