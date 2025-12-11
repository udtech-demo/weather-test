package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

func InitRedis() *redis.Client {
	//Initializing redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     viper.GetString("rdb.host") + viper.GetString("rdb.port"),
		Password: viper.GetString("rdb.pwd"),
		DB:       viper.GetInt("rdb.db"),
	})

	ctx := context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}

	//rdb.FlushAll(ctx)

	return rdb
}
