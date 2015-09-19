/**
 * @file pulse.go
 * @author Mikhail Klementyev jollheef<AT>riseup.net
 * @license GNU GPLv3
 * @date September, 2015
 * @brief generate game events
 *
 * Routine that generates game events
 */

package pulse

import (
	"crypto/rsa"
	"database/sql"
	"log"
	"time"
)

func Wait(end time.Time, timeout time.Duration) (waited bool) {

	if time.Now().After(end) {
		return false
	}

	for time.Now().Before(end) {
		time.Sleep(timeout)
	}

	return true
}

func Pulse(db *sql.DB, priv *rsa.PrivateKey, start_time time.Time,
	half, lunch, round_len, check_timeout time.Duration) (err error) {

	log.Println("Launching pulse...")

	lunch_start_time := start_time.Add(half)
	lunch_end_time := lunch_start_time.Add(lunch)
	end_time := lunch_end_time.Add(half)

	log.Println("Pulse start time", time.Now())

	log.Println("Contest start time", start_time)

	game, err := NewGame(db, priv, round_len, check_timeout)

	defer game.Over()

	timeout := 100 * time.Millisecond

	log.Println("Wait start time...")
	if Wait(start_time, timeout) || time.Now().Before(lunch_start_time) {
		log.Println("game run")
		err = game.Run(lunch_start_time)
		if err != nil {
			return
		}
	}

	log.Println("Wait lunch time")
	Wait(lunch_start_time, timeout)

	log.Println("Lunch...")
	if Wait(lunch_end_time, timeout) || time.Now().Before(end_time) {
		err = game.Run(end_time)
		if err != nil {
			return
		}
	}

	log.Println("Wait end time")
	Wait(end_time, timeout)

	return
}