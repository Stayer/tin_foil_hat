/**
 * @file main.go
 * @author Mikhail Klementyev jollheef<AT>riseup.net
 * @license GNU AGPLv3
 * @date September, 2015
 * @brief contest checking system CLI
 *
 * Entry point for contest checking system CLI
 */

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/olekukonko/tablewriter"
	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"github.com/jollheef/tin_foil_hat/config"
	"github.com/jollheef/tin_foil_hat/scoreboard"
	"github.com/jollheef/tin_foil_hat/steward"
)

var (
	configPath = kingpin.Flag("config",
		"Path to configuration file.").String()

	score = kingpin.Command("scoreboard", "View scoreboard.")

	adv = kingpin.Command("advisory", "Work with advisories.")

	advList        = adv.Command("list", "List advisories.")
	advNotReviewed = adv.Flag("not-reviewed",
		"List only not reviewed advisory.").Bool()

	advReview   = adv.Command("review", "Review advisory.")
	advReviewID = advReview.Arg("id", "advisory id").Required().Int()
	advScore    = advReview.Arg("score", "advisory id").Required().Int()

	advHide   = adv.Command("hide", "Hide advisory.")
	advHideID = advHide.Arg("id", "advisory id").Required().Int()

	advUnhide   = adv.Command("unhide", "Unhide advisory.")
	advUnhideID = advUnhide.Arg("id", "advisory id").Required().Int()
)

var (
	commitID  string
	buildDate string
	buildTime string
)

func buildInfo() (str string) {

	if len(commitID) > 7 {
		commitID = commitID[:7] // abbreviated commit hash
	}

	str = fmt.Sprintf("Version: tin_foil_hat %s %s %s\n",
		commitID, buildDate, buildTime)
	str += "Author: Mikhail Klementyev <jollheef@riseup.net>\n"
	return
}

func advisoryList(db *sql.DB) {
	advisories, err := steward.GetAdvisories(db)
	if err != nil {
		log.Fatalln("Get advisories fail:", err)
	}

	for _, advisory := range advisories {

		if *advNotReviewed && advisory.Reviewed {
			continue
		}

		fmt.Printf(">>> Advisory: id %d <<<\n", advisory.ID)
		fmt.Printf("(Score: %d, Reviewed: %t, Timestamp: %s)\n",
			advisory.Score, advisory.Reviewed,
			advisory.Timestamp.String())

		fmt.Println(advisory.Text)
	}

}

func advisoryReview(db *sql.DB) {
	err := steward.ReviewAdvisory(db, *advReviewID, *advScore)
	if err != nil {
		log.Fatalln("Advisory review fail:", err)
	}
}

func advisoryHide(db *sql.DB) {
	err := steward.HideAdvisory(db, *advHideID, true)
	if err != nil {
		log.Fatalln("Advisory hide fail:", err)
	}

}

func advisoryUnhide(db *sql.DB) {
	err := steward.HideAdvisory(db, *advUnhideID, false)
	if err != nil {
		log.Fatalln("Advisory unhide fail:", err)
	}
}

func scoreboardShow(db *sql.DB) {
	res, err := scoreboard.CollectLastResult(db)
	if err != nil {
		log.Fatalln("Get last result fail:", err)
	}

	scoreboard.CountScoreAndSort(&res)

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Rank", "Name", "Score", "Attack",
		"Defence", "Advisory"})

	for _, tr := range res.Teams {

		var row []string

		row = append(row, fmt.Sprintf("%d", tr.Rank))
		row = append(row, tr.Name)
		row = append(row, fmt.Sprintf("%05.2f%%", tr.ScorePercent))
		row = append(row, fmt.Sprintf("%.3f", tr.Attack))
		row = append(row, fmt.Sprintf("%.3f", tr.Defence))
		row = append(row, fmt.Sprintf("%d", tr.Advisory))

		table.Append(row)
	}

	table.Render()
}

func main() {

	fmt.Println(buildInfo())

	kingpin.Parse()

	if *configPath == "" {
		*configPath = "/etc/tinfoilhat/tinfoilhat.toml"
	}

	config, err := config.ReadConfig(*configPath)
	if err != nil {
		log.Fatalln("Cannot open config:", err)
	}

	db, err := steward.OpenDatabase(config.Database.Connection)
	if err != nil {
		log.Fatalln("Open database fail:", err)
	}

	defer db.Close()

	db.SetMaxOpenConns(config.Database.MaxConnections)

	switch kingpin.Parse() {
	case "advisory list":
		advisoryList(db)

	case "advisory review":
		advisoryReview(db)

	case "advisory hide":
		advisoryHide(db)

	case "advisory unhide":
		advisoryUnhide(db)

	case "scoreboard":
		scoreboardShow(db)
	}
}
