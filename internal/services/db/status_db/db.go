package statusdb

import (
	"log"

	"github.com/go-redis/redis"
)

func Connect() {
	opt, err := redis.ParseURL("redis://statusdb-ree9or.serverless.eun1.cache.amazonaws.com:6379")
	if err != nil {
		log.Fatal(err.Error())
	}

	client := redis.NewClient(opt)
	err = client.Set("foo", "bar", 0).Err()
	if err != nil {
		log.Fatal(err.Error())
	}

	res, err := client.Get("foo").Result()
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Println("foo", res)

}
