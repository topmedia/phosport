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

type Rule struct {
	Data         RuleData `json:"data"`
	Destinations []string `json:"-"`
	Sources      []string `json:"-"`
	Services     []string `json:"-"`
}

func (r *Rule) ResolveRefs() {
	for _, src := range r.Data.Sources {
		var host Host
		ResolveRef(src, &host)
		for _, h := range host.MembersAndSelf() {
			r.Sources = append(r.Sources, h.Address())
		}
	}

	for _, dst := range r.Data.Destinations {
		var host Host
		ResolveRef(dst, &host)
		for _, h := range host.MembersAndSelf() {
			r.Destinations = append(r.Destinations, h.Address())
		}
	}

	for _, sv := range r.Data.Services {
		var svc Service
		ResolveRef(sv, &svc)
		for _, h := range svc.MembersAndSelf() {
			r.Services = append(r.Services, h.Ports())
		}
	}
}

type RuleData struct {
	Action       string   `json:"action"`
	Comment      string   `json:"comment"`
	Destinations []string `json:"destinations"`
	Group        string   `json:"group"`
	Interface    string   `json:"interface"`
	Name         string   `json:"name"`
	Services     []string `json:"services"`
	Sources      []string `json:"sources"`
	Status       int      `json:"status"`
}

type Host struct {
	Data  HostData `json:"data"`
	Class string   `json:"class"`
	Type  string   `json:"type"`
}

func (h *Host) Address() string {
	if h.Data.Address == "" {
		return h.Data.Name
	} else if h.Data.Netmask != 0 {
		return fmt.Sprintf("%s/%d", h.Data.Address, h.Data.Netmask)
	} else {
		return h.Data.Address
	}
}

func (h *Host) MembersAndSelf() (hosts []Host) {
	if h.Type != "group" {
		return append(hosts, *h)
	}

	for _, m := range h.Data.Members {
		var h Host
		ResolveRef(m, &h)
		hosts = append(hosts, h)
	}

	return hosts
}

type HostData struct {
	Address string   `json:"address"`
	Comment string   `json:"comment"`
	Name    string   `json:"name"`
	Netmask int      `json:"netmask"`
	Members []string `json:"members"`
}

type Service struct {
	Data  ServiceData `json:"data"`
	Class string      `json:"class"`
	Type  string      `json:"type"`
}

func (s *Service) Ports() string {
	if s.Data.DstHigh == 0 {
		return "1:65535"
	} else if s.Data.DstHigh == s.Data.DstLow {
		return fmt.Sprintf("%d", s.Data.DstLow)
	} else {
		return fmt.Sprintf("%d:%d", s.Data.DstLow, s.Data.DstHigh)
	}
}

func (s *Service) MembersAndSelf() (services []Service) {
	if s.Type != "group" {
		return append(services, *s)
	}

	for _, m := range s.Data.Members {
		var s Service
		ResolveRef(m, &s)
		services = append(services, s)
	}

	return services
}

type ServiceData struct {
	DstHigh int      `json:"dst_high"`
	DstLow  int      `json:"dst_low"`
	Comment string   `json:"comment"`
	Name    string   `json:"name"`
	Members []string `json:"members"`
}

type RulePrint struct {
	Sources      []string `json:"sources"`
	Destinations []string `json:"destinations"`
	Services     []string `json:"services"`
}

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
