package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var (
	host = flag.String("host", "localhost",
		"UTM Hostname")
	verbose = flag.Bool("v", false,
		"Output all commands executed on UTM")
	resolve = flag.Bool("resolve", false,
		"When exporting rules, resolve all the REFs")
	rules = flag.Bool("rules", false,
		"This output mode exports all packetfilter rules")
	hostgroups = flag.Bool("hostgroups", false,
		"This output mode exports all host groups")
	hosts = flag.Bool("hosts", false,
		"This output mode exports all hosts")
	services = flag.Bool("services", false,
		"This output mode exports all services")
	servicegroups = flag.Bool("servicegroups", false,
		"This output mode exports all service groups")
)

const cc = "confd-client.plx"

// Convert the confd-client output to "real" JSON
// with double quotes and colons
func ToJSON(input []byte) []byte {
	vars := regexp.MustCompile(`(\$VAR[^,]+),`)
	fixquotes := func(m string) string {
		return fmt.Sprintf(`["%s"],`,
			strings.Replace(m, `"`, "'", -1))
	}

	str := strings.Replace(string(input), " => ", ": ", -1)
	str = strings.Replace(str, `"`, `\"`, -1)
	str = strings.Replace(str, "'", `"`, -1)
	str = vars.ReplaceAllStringFunc(str, fixquotes)
	return []byte(str)
}

// Execute a confd-client command either locally or remotely
func ConfdCommand(cmds ...string) (out []byte) {
	if *verbose {
		log.Printf("Executing command %v on host %s", cmds, *host)
	}

	cmd := exec.Command(cc, cmds...)

	if *host != "localhost" {
		if len(cmds) > 1 {
			cmds[1] = fmt.Sprintf("'%s'", cmds[1])
		}
		cmds = append([]string{*host, cc}, cmds...)
		cmd = exec.Command("ssh", cmds...)
	}

	out, err := cmd.Output()

	if err != nil {
		log.Fatalf("Error executing confd command: %s %s", err, out)
	}

	return out
}

// Resolves a REF_ string to an object
func ResolveRef(refstr string, target interface{}) {
	if strings.HasPrefix(refstr, "$VAR") {
		return
	}

	ref := ConfdCommand("get_object", refstr)
	err := json.Unmarshal(ToJSON([]byte(ref)), &target)

	if err != nil {
		log.Fatalf("Error resolving REF %s: %v", refstr, err)
	}
}

func OutputJSON(objs interface{}) {
	out, err := json.MarshalIndent(objs, "", "  ")

	if err != nil {
		log.Fatalf("Error preparing output JSON: %v", err)
	}

	os.Stdout.Write(out)
}

func main() {
	flag.Parse()

	if *rules {
		OutputJSON(ExportRules(*resolve))
	} else if *hostgroups {
		OutputJSON(ExportGroups("network"))
	} else if *hosts {
		OutputJSON(ExportHosts())
	} else if *servicegroups {
		OutputJSON(ExportGroups("service"))
	} else if *services {
		OutputJSON(ExportServices())
	} else {
		fmt.Println("Please choose a mode to export.\n\nUsage:")
		flag.PrintDefaults()
	}
}
