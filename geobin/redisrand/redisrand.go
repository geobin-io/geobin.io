package redisrand

import (
	redis "github.com/vmihailenco/redis/v2"
	"math/rand"
	"time"
)

const stringVals = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func init() {
	// set up seed for our random number generator
	rand.Seed(time.Now().UTC().UnixNano())
}

type RedisRand interface {
	NextString(handler RedisRandHandler, length int)
}

type RedisRandHandler func(randomString string)

type redisRand struct {
	reqs chan *randRequest
}

type randRequest struct {
	handler RedisRandHandler
	length int
}

func NewRedisRand(redisHost, redisPass string, redisDB int64) RedisRand {
	// a channel for getting a new random string
	reqs := make(chan *randRequest)

	rr := &redisRand{
		reqs,
	}

	go rr.handleRands(redisHost, redisPass, redisDB)

	return rr
}

func (rs *redisRand) NextString(handler RedisRandHandler, length int) {
	go func() {
		rs.reqs <- &randRequest{handler, length}
	}()
}

func (rs *redisRand) handleRands(redisHost, redisPass string, redisDB int64) {
	for {
		req := <-rs.reqs

		// connect to redis
		client := redis.NewTCPClient(&redis.Options{
			Addr:     redisHost,
			Password: redisPass,
			DB:       redisDB,
		})

		req.handler(randomString(client, req.length))

		client.Close()
	}
}

func randomString(client *redis.Client, length int) string {
	newBytes := make([]byte, length)
	for i, _ := range newBytes {
		newBytes[i] = stringVals[rand.Intn(len(stringVals))]
	}

	newString := string(newBytes)
	if client.Exists(newString).Val() {
		return randomString(client, length)
	}

	return newString
}
