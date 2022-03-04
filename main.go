package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	return strings.Join(*i, "")
}
func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}
func waitFor(waitsFlags arrayFlags, commandFlags string, timeoutFlag int, intervalFlag int) {
	for _, wait := range waitsFlags {
		processWait(wait, timeoutFlag, intervalFlag)
	}
	processCommandExec(commandFlags)
}

func processWait(wait string, timeoutFlag, intervalFlag int) {
	pattern := regexp.MustCompile("(.*):([0-9]+)")
	matches := pattern.FindAllStringSubmatch(wait, -1)
	startTime := time.Now()

	dbHost := matches[0][1]
	dbPort := matches[0][2]
	for {
		d := net.Dialer{Timeout: time.Duration(intervalFlag) * time.Second}
		ctx := context.Background()
		ctx, _ = context.WithTimeout(ctx, time.Duration(timeoutFlag)*time.Second)
		conn, err := d.DialContext(ctx, "tcp", net.JoinHostPort(dbHost, dbPort))
		if err != nil {
			fmt.Printf("Sleeping %d milliseconds waiting for %v\n", intervalFlag, net.JoinHostPort(dbHost, dbPort))
			time.Sleep(time.Duration(intervalFlag) * time.Millisecond)
		}
		if conn != nil {
			fmt.Printf("%v reading in %v\n", net.JoinHostPort(dbHost, dbPort), time.Now().Sub(startTime).String())
			_ = conn.Close()
			break
		}
	}

}

func processCommandExec(command string) {
	cmdArgs := strings.Split(command, " ")
	cmd := cmdArgs[0]
	args := cmdArgs[1:]

	out := exec.Command(cmd, args...)
	out.Stdout = os.Stdout
	out.Stderr = os.Stderr
	out.Stdin = os.Stdin
	err := out.Run()

	if err != nil {
		log.Fatal(err)
	}
}

var (
	waitsFlags   arrayFlags
	commandFlags arrayFlags
	timeoutFlag  *int
	intervalFlag *int
)

func init() {
	timeoutFlag = flag.Int("timeout", 60, "Timeout until script is killed.")
	intervalFlag = flag.Int("interval", 250, "Interval between calls")
	flag.Var(&waitsFlags, "wait", "You can specify the HOST and TCP PORT using the format HOST:PORT, or you can specify a command that should return an output. Multiple wait flags can be added.")
	flag.Var(&commandFlags, "command", "Command that should be run when all waits are accessible. Multiple commands can be added.")
	flag.Parse()
}
func main() {
	Exec()
}

func Exec() {
	if len(waitsFlags) == 0 || len(commandFlags) == 0 {
		fmt.Println("You must specify at least a wait and a command. Please see --help for more information.")
		return
	}
	waitFor(waitsFlags, commandFlags.String(), *timeoutFlag, *intervalFlag)
}
