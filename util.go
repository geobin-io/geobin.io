package main

import (
	"log"
	"math/rand"
)

// randomString returns a random string with the given length
func randomString(length int) (string, error) {
	b := make([]byte, length)
	for i := range b {
		b[i] = config.NameVals[rand.Intn(len(config.NameVals))]
	}

	s := string(b)

	exists, err := nameExists(s)
	if err != nil {
		log.Println("Failure to EXISTS for:", s, err)
		return "", err
	}

	if exists {
		return randomString(length)
	}

	return s, nil
}

// nameExists returns true if the specified bin name exists
func nameExists(name string) (bool, error) {
	resp := client.Exists(name)
	if resp.Err() != nil {
		return false, resp.Err()
	}

	return resp.Val(), nil
}

func debugLog(v ...interface{}) {
	if *isDebug {
		log.Println(v...)
	}
}
