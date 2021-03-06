/**
 * @file checker.go
 * @author Mikhail Klementyev jollheef<AT>riseup.net
 * @license GNU AGPLv3
 * @date September, 2015
 * @brief functions for check services
 *
 * Provide functions for check service status, put flags and check flags.
 */

package checker

import (
	"crypto/rsa"
	"database/sql"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/jollheef/tin_foil_hat/steward"
	"github.com/jollheef/tin_foil_hat/vexillary"
)

func tcpPortOpen(team steward.Team, svc steward.Service) bool {

	addr := fmt.Sprintf("%s:%d", team.Vulnbox, svc.Port)

	conn, err := net.DialTimeout("tcp", addr, portCheckTimeout)
	if err != nil {
		return false
	}

	conn.Close()
	return true
}

func putFlag(db *sql.DB, priv *rsa.PrivateKey, round int, team steward.Team,
	svc steward.Service) (err error) {

	flag, err := vexillary.GenerateFlag(priv)
	if err != nil {
		log.Println("Generate flag failed:", err)
		return
	}

	portOpen := true
	if !svc.UDP {
		portOpen = tcpPortOpen(team, svc)
	}

	var cred, logs string
	var state steward.ServiceState
	if portOpen {
		if team.UseNetbox {
			cred, logs, state, err = sshPut(team.Netbox,
				svc.CheckerPath, team.Vulnbox, svc.Port, flag)
		} else {
			cred, logs, state, err = put(svc.CheckerPath,
				team.Vulnbox, svc.Port, flag)
		}
		if err != nil {
			log.Println("Put flag to service failed:", err)
			return
		}

		if state != steward.StatusUP {
			log.Printf("Put flag, round %d, team %s, service %s: %s",
				round, team.Name, svc.Name, logs)
		}
	} else {
		state = steward.StatusDown
	}

	err = steward.PutStatus(db,
		steward.Status{round, team.ID, svc.ID, state})
	if err != nil {
		log.Println("Add status to database failed:", err)
		return
	}

	err = steward.AddFlag(db,
		steward.Flag{-1, flag, round, team.ID, svc.ID, cred})
	if err != nil {
		log.Println("Add flag to database failed:", err)
		return
	}

	return
}

func getFlag(db *sql.DB, round int, team steward.Team,
	svc steward.Service) (state steward.ServiceState, err error) {

	flag, cred, err := steward.GetCred(db, round, team.ID, svc.ID)
	if err != nil {
		log.Println("Get cred failed:", err)
		state = steward.StatusCorrupt
		return
	}

	var logs string
	var serviceFlag string

	if team.UseNetbox {
		serviceFlag, logs, state, err = sshGet(team.Netbox,
			svc.CheckerPath, team.Vulnbox, svc.Port, cred)
	} else {
		serviceFlag, logs, state, err = get(svc.CheckerPath,
			team.Vulnbox, svc.Port, cred)
	}
	if err != nil {
		log.Println("Check service failed:", err)
		return
	}

	if flag != serviceFlag {
		state = steward.StatusCorrupt
	}

	if state != steward.StatusUP {
		log.Printf("Get flag, round %d, team %s, service %s: %s",
			round, team.Name, svc.Name, logs)
	}

	return
}

func checkService(db *sql.DB, round int, team steward.Team,
	svc steward.Service) (state steward.ServiceState, err error) {

	var logs string

	if team.UseNetbox {
		state, logs, err = sshCheck(team.Netbox, svc.CheckerPath,
			team.Vulnbox, svc.Port)
	} else {
		state, logs, err = check(svc.CheckerPath, team.Vulnbox,
			svc.Port)
	}
	if err != nil {
		log.Println("Check service failed:", err)
		return
	}

	if state != steward.StatusUP {
		log.Printf("Check, round %d, team %s, service %s: %s",
			round, team.Name, svc.Name, logs)
	}

	return
}

// Check service status and flag if it's exist.
func checkFlag(db *sql.DB, round int, team steward.Team, svc steward.Service,
	wg *sync.WaitGroup) {

	defer wg.Done()

	// Check service port open
	portOpen := true
	if !svc.UDP {
		portOpen = tcpPortOpen(team, svc)
	}

	var state steward.ServiceState
	if portOpen {
		// First check service logic
		state, _ = checkService(db, round, team, svc)
		if state == steward.StatusUP {
			// If logic is correct, do flag check
			state, _ = getFlag(db, round, team, svc)
		}
	} else {
		state = steward.StatusDown
	}

	err := steward.PutStatus(db, steward.Status{round,
		team.ID, svc.ID, state})
	if err != nil {
		log.Println("Add status failed:", err)
		return
	}
}

// PutFlags put flags to services
func PutFlags(db *sql.DB, priv *rsa.PrivateKey, round int,
	teams []steward.Team, services []steward.Service) (err error) {

	var wg sync.WaitGroup

	for _, team := range teams {
		for _, svc := range services {
			wg.Add(1)
			go func(team steward.Team, svc steward.Service) {
				defer wg.Done()
				putFlag(db, priv, round, team, svc)
			}(team, svc)
		}
	}

	wg.Wait()

	return
}

// CheckFlags check flags in services
func CheckFlags(db *sql.DB, round int, teams []steward.Team,
	services []steward.Service) (err error) {

	var wg sync.WaitGroup

	for _, team := range teams {
		for _, svc := range services {
			wg.Add(1)
			go checkFlag(db, round, team, svc, &wg)
		}
	}

	wg.Wait()

	return
}
