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

func main() {
	flag.Parse()

	pkf := ConfdCommand("get_objects_filtered", `$_->{type} eq "packetfilter"`)

	var rules []Rule
	err := json.Unmarshal(ToJSON(pkf), &rules)
	if err != nil {
		log.Fatalf("Error parsing rules into JSON: %s", err)
	}

	if *verbose {
		log.Printf("Found %d rules, resolving objects",
			len(rules))
	}
	rulesprint := make([]RulePrint, len(rules))

	for _, rule := range rules {
		rule.ResolveRefs()
		rulesprint = append(rulesprint, RulePrint{
			Sources:      rule.Sources,
			Destinations: rule.Destinations,
			Services:     rule.Services,
		})
	}

	out, err := json.MarshalIndent(rulesprint, "", "  ")

	if err != nil {
		log.Fatalf("Error preparing output JSON: %v", err)
	}

	os.Stdout.Write(out)
}
