package conf

import (
	"github.com/spf13/viper"
)

type Conf struct {
}

func Init() error {
	viper.AddConfigPath("./conf")
	viper.SetConfigName("config")

	return viper.ReadInConfig()
}
