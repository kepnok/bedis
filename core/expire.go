package core

import (
	"log"
	"time"
)

func expire() float32 {
	totalChecks := 20
	totalDeleted := 0

	for key, obj := range store {
		if obj.ExpiresAt != -1 {
			totalChecks--

			if obj.ExpiresAt <= time.Now().UnixMilli() {
				delete(store, key)
				totalDeleted++
			}
		}	

		if totalChecks == 0 {
			break
		}
	}

	return float32(totalDeleted)/float32(20)
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