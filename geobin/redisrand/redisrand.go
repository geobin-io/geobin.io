package redisrand

import (
	redis "github.com/vmihailenco/redis/v2"
	"math/rand"
	"time"
)

const length = 12

const stringVals = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func init() {
	// set up seed for our random number generator
	rand.Seed(time.Now().UTC().UnixNano())
}

type RedisRand interface {
	NextString(*randRequest)
}

type randRequest struct {
	StringResponse chan string
}

type redisRand struct {
	reqs chan *randRequest
}

func NewRedisRand(redisHost, redisPass string, redisDB int64) RedisRand {
	// a channel for getting a new random string
	reqs := make(chan *randRequest)

	rr := &redisRand{
		reqs,
	}

	go rr.manageStrings(redisHost, redisPass, redisDB)

	return rr
}

func (rs *redisRand) NextString(req *randRequest) {
	go func() {
		rs.reqs <- req
	}()
}

func (rs *redisRand) manageStrings(redisHost, redisPass string, redisDB int64) {
	for {
		req := <-rs.reqs

		// connect to redis
		client := redis.NewTCPClient(&redis.Options{
			Addr:     redisHost,
			Password: redisPass,
			DB:       redisDB,
		})

		rando := randomString(client)

		go func(rando string) {
			req.StringResponse <- rando
		}(rando)

		client.Close()
	}
}

func randomString(client *redis.Client) string {
	newBytes := make([]byte, length)
	for i, _ := range newBytes {
		newBytes[i] = stringVals[rand.Intn(len(stringVals))]
	}

	newString := string(newBytes)
	if client.Exists(newString).Val() {
		return randomString(client)
	}

	client.ZAdd(newString, Z{
		-1,
		"placeholder",
	})

	return newString
}
