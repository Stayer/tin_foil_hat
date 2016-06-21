/**
 * @file status_test.go
 * @author Mikhail Klementyev jollheef<AT>riseup.net
 * @license GNU AGPLv3
 * @date September, 2015
 * @brief test work with status table
 */

package steward_test

import (
	"log"
	"testing"
)

import "github.com/jollheef/tin_foil_hat/steward"

func TestPutStatus(t *testing.T) {

	db, err := openDB()

	defer db.Close()

	status := steward.Status{10, 10, 10, 10}

	err = steward.PutStatus(db.db, status)
	if err != nil {
		log.Fatalln("Add status failed:", err)
	}
}

func TestGetAllStatus(t *testing.T) {

	db, err := openDB()

	defer db.Close()

	round := 1
	team := 2
	service := 3

	status1 := steward.Status{round, team, service, steward.STATUS_UP}
	status2 := steward.Status{round, team, service, steward.STATUS_MUMBLE}
	status3 := steward.Status{round, team, service, steward.STATUS_CORRUPT}

	steward.PutStatus(db.db, status1)
	steward.PutStatus(db.db, status2)
	steward.PutStatus(db.db, status3)

	halfStatus := steward.Status{round, team, service,
		steward.STATUS_UNKNOWN}

	states, err := steward.GetStates(db.db, halfStatus)
	if err != nil {
		log.Fatalln("Get states failed:", err)
	}

	if len(states) != 3 {
		log.Fatalln("Get states moar/less than put:", err)
	}

	if states[0] != steward.STATUS_UP ||
		states[1] != steward.STATUS_MUMBLE ||
		states[2] != steward.STATUS_CORRUPT {
		log.Fatalln("Get states invalid")
	}
}

func TestGetServiceCurrentStatus(t *testing.T) {

	db, err := openDB()

	defer db.Close()

	round := 1
	team := 2
	service := 3

	status1 := steward.Status{round, team, service, steward.STATUS_UP}

	steward.PutStatus(db.db, status1)

	halfStatus := steward.Status{round, team, service,
		steward.STATUS_UNKNOWN}

	state, err := steward.GetState(db.db, halfStatus)
	if err != nil {
		log.Fatalln("Get state failed:", err)
	}

	if state != steward.STATUS_UP {
		log.Fatalln("Get states invalid")
	}

}