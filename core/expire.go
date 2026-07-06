package core

import (
	"log"
)

func expire() float32 {
	totalChecks := 20
	totalDeleted := 0

	for key, obj := range store {

		totalChecks--

		if hasExpired(obj) {
			Del(key)
			totalDeleted++
		}

		if totalChecks == 0 {
			break
		}
	}

	return float32(totalDeleted) / float32(20)
}

func DeleteExpiredKeys() {
	for {
		frac := expire()

		if frac < 0.25 {
			break
		}
	}

	log.Println("deleted the undeleted expired keys. Total keys left: ", len(store))
}