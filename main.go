package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

type Packetfilter struct {
	Action       string
	Comment      string
	Destinations []string
	Group        string
	Interface    string
	Name         string
	Services     []string
	Sources      []string
	Status       bool
}

var (
	host = flag.String("host", "192.168.1.1",
		"UTM Hostname")
)

func ToJSON(input string) string {

	input = strings.Replace(input, "'", `"`, -1)
	input = strings.Replace(input, " => ", ": ", -1)
	return input
}

func main() {
	flag.Parse()

	pkf, err := exec.Command("ssh", *host, "confd-client.plx",
		`get_objects_filtered '$_->{type} eq "packetfilter"'`).
		Output()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(ToJSON(string(pkf)))
}
