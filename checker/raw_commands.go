/**
 * @file raw_commands.go
 * @author Mikhail Klementyev jollheef<AT>riseup.net
 * @license GNU AGPLv3
 * @date September, 2015
 * @brief functions for run checkers
 *
 * Provide functions for call checker executables
 */

package checker

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
	"time"
)

import "github.com/jollheef/tin_foil_hat/steward"

var timeout = "10s" // max checker work time

// SetTimeout set max checker work time
func SetTimeout(d time.Duration) {
	timeout = fmt.Sprintf("%ds", int(d.Seconds()))
}

func readBytesUntilEOF(pipe io.ReadCloser) (buf []byte, err error) {

	bufSize := 1024

	for err != io.EOF {
		stdout := make([]byte, bufSize)
		var n int

		n, err = pipe.Read(stdout)
		if err != nil && err != io.EOF {
			return
		}

		buf = append(buf, stdout[:n]...)
	}

	if err == io.EOF {
		err = nil
	}

	return
}

func readUntilEOF(pipe io.ReadCloser) (str string, err error) {
	buf, err := readBytesUntilEOF(pipe)
	str = string(buf)
	return
}

func system(name string, arg ...string) (stdout string, stderr string,
	err error) {

	cmd := exec.Command(name, arg...)

	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		return
	}

	errPipe, err := cmd.StderrPipe()
	if err != nil {
		return
	}

	cmd.Start()

	stdout, err = readUntilEOF(outPipe)
	if err != nil {
		return
	}

	stderr, err = readUntilEOF(errPipe)
	if err != nil {
		return
	}

	err = cmd.Wait()

	return
}

func exitStatus(no int) string {
	return fmt.Sprintf("exit status %d", no)
}

func parseState(err error) (steward.ServiceState, error) {

	if err == nil {
		return steward.STATUS_UP, nil
	}

	switch err.Error() {
	case exitStatus(124): // returns by timeout
		return steward.STATUS_DOWN, nil
	case exitStatus(1):
		return steward.STATUS_ERROR, nil
	case exitStatus(2):
		return steward.STATUS_MUMBLE, nil
	case exitStatus(3):
		return steward.STATUS_CORRUPT, nil
	case exitStatus(4):
		return steward.STATUS_DOWN, nil
	}

	return steward.STATUS_UNKNOWN, err
}

func put(checker string, ip string, port int, flag string) (cred, logs string,
	state steward.ServiceState, err error) {

	cred, logs, err = system("timeout", timeout, checker, "put", ip,
		fmt.Sprintf("%d", port), flag)

	state, err = parseState(err)

	cred = strings.Trim(cred, " \n")

	return
}

func get(checker string, ip string, port int, cred string) (flag, logs string,
	state steward.ServiceState, err error) {

	flag, logs, err = system("timeout", timeout, checker, "get", ip,
		fmt.Sprintf("%d", port), cred)

	state, err = parseState(err)

	flag = strings.Trim(flag, " \n")

	return
}

func check(checker string, ip string, port int) (state steward.ServiceState,
	logs string, err error) {

	_, logs, err = system("timeout", timeout, checker, "chk", ip,
		fmt.Sprintf("%d", port))

	state, err = parseState(err)

	return
}