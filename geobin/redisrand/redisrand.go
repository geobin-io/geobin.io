package redisrand

import (
	redis "github.com/vmihailenco/redis/v2"
	"math/rand"
	"time"
)

const length = 12

const stringVals = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type RedisRand interface {
	NextString(randRequest)
}

type randRequest struct {
	StringResponse chan string
}

type redisRand struct {
	reqs chan *randRequest
	client *redis.Client
}

func NewRedisRand(redisHost, redisPass string, redisDB int64) RedisRand {
	// set up seed for our random number generator
	rand.Seed(time.Now().UTC().UnixNano())

	// connect to redis
	client := redis.NewTCPClient(&redis.Options{
		Addr:     redisHost,
		Password: redisPass,
		DB:       redisDB,
	})
	// a channel for getting a new random string
	reqs := make(chan *randRequest)

	rr := &redisRand {
		reqs,
		client,
	}

	go rr.manageStrings()

	return rr
}

func (rs *redisRand) NextString(req *randRequest) string {
	go func() {
		rs.reqs <- req
	}()
}

func (rs *redisRand) manageStrings() {
	defer rs.redisClient.Close()
	for {
		req := <- rs.reqs
		rando := rs.randomString()

		go func(newString string) {
			req <- newString
		}(rando)
	}
}

func (rs *redisRand) randomString() string {
	newString := make([]byte, length, length)
	for i, _ := range newString {
		newString[i] = stringVals[rand.Intn(len(stringVals) - 1)]
	}

	if rs.client.Exists(string(newString)) {
		return rs.randomString()
	}

	return string(newString)
}
