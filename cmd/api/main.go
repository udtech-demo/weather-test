package main

import (
	"weather-data-aggregator-service/conf"
	"weather-data-aggregator-service/src/server"
)

func main() {
	if err := conf.Init(); err != nil {
		panic(err)
	}

	app := server.NewApp()

	if err := app.Run(); err != nil {
		panic(err)
	}
}
