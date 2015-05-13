package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"
)

type Rule struct {
	Data RuleData `json:"data"`
}

func (r *Rule) ResolveRefs() {
	for i, src := range r.Data.Sources {
		var host Host
		ResolveRef(src, &host)
		r.Data.Sources[i] = host.Address()
	}

	for i, dst := range r.Data.Destinations {
		var host Host
		ResolveRef(dst, &host)
		r.Data.Destinations[i] = host.Address()
	}

	for i, sv := range r.Data.Services {
		var svc Service
		ResolveRef(sv, &svc)
		r.Data.Services[i] = svc.Ports()
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
}

func (h *Host) Address() string {
	if h.Data.Address == "" {
		return h.Data.Name
	} else {
		return h.Data.Address
	}
}

type HostData struct {
	Address string `json:"address"`
	Comment string `json:"comment"`
	Name    string `json:"name"`
}

type Service struct {
	Data  ServiceData `json:"data"`
	Class string      `json:"class"`
}

func (s *Service) Ports() string {
	if s.Data.DstHigh == 0 {
		return "1:65535"
	} else {
		return fmt.Sprintf("%d:%d", s.Data.DstLow, s.Data.DstHigh)
	}
}

type ServiceData struct {
	DstHigh int    `json:"dst_high"`
	DstLow  int    `json:"dst_low"`
	Comment string `json:"comment"`
	Name    string `json:"name"`
}

var (
	host = flag.String("host", "192.168.1.1",
		"UTM Hostname")
)

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

func ConfdCommand(cmd string) []byte {
	log.Printf("Executing command %s on host %s", cmd, *host)
	out, err := exec.Command("ssh", *host, "confd-client.plx",
		cmd).Output()

	if err != nil {
		log.Fatalf("Error executing confd command: %s", err)
	}

	return out
}

func ResolveRef(refstr string, target interface{}) {
	if strings.HasPrefix(refstr, "$VAR") {
		return
	}

	ref := ConfdCommand(fmt.Sprintf("get_object %s", refstr))
	err := json.Unmarshal(ToJSON(ref), &target)
	if err != nil {
		log.Fatalf("Error resolving REF: %s", refstr, err)
	}
}

func main() {
	flag.Parse()

	pkf := ConfdCommand(`get_objects_filtered '$_->{type} eq "packetfilter"'`)

	var rules []Rule
	err := json.Unmarshal(ToJSON(pkf), &rules)
	if err != nil {
		log.Fatalf("Error parsing rules into JSON: %s", err)
	}

	for _, rule := range rules {
		rule.ResolveRefs()
		log.Printf("Source: %v | Dest: %v | Services: %v", rule.Data.Sources, rule.Data.Destinations, rule.Data.Services)
	}

	// log.Printf("Rules: %#v", rules)

}
