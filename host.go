package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

type Host struct {
	Data  HostData `json:"data"`
	Class string   `json:"class"`
	Type  string   `json:"type"`
	Ref   string   `json:"ref"`
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
	HostDataPrint
	Members []string `json:"members"`
}

type HostDataPrint struct {
	Address   string   `json:"address"`
	Comment   string   `json:"comment"`
	Hostnames []string `json:"hostnames"`
	Macs      []string `json:"macs"`
	Name      string   `json:"name"`
	Netmask   int      `json:"netmask"`
	Interface string   `json:"interface"`
	Ref       string   `json:"ref"`
}

func (hd *HostData) MarshalJSON() ([]byte, error) {
	hdp := HostDataPrint{
		Address:   hd.Address,
		Comment:   hd.Comment,
		Name:      hd.Name,
		Netmask:   hd.Netmask,
		Interface: hd.Interface,
		Ref:       hd.Ref,
	}

	if hns := hd.Hostnames; len(hns) > 0 {
		if !strings.HasPrefix(hns[0], "$VAR") {
			hdp.Hostnames = hd.Hostnames
		}
	}

	if macs := hd.Macs; len(macs) > 0 {
		if !strings.HasPrefix(macs[0], "$VAR") {
			hdp.Macs = hd.Macs
		}
	}

	return json.Marshal(hdp)
}

func ExportHosts() []HostData {
	objs := ConfdCommand("get_objects_filtered", `$_->{type} eq "host"`)

	var hosts []Host
	err := json.Unmarshal(ToJSON(objs), &hosts)
	if err != nil {
		log.Fatalf("Error parsing groups into JSON: %s", err)
	}

	hostsprint := make([]HostData, len(hosts))

	for _, host := range hosts {
		host.Data.Ref = host.Ref
		hostsprint = append(hostsprint, host.Data)
	}
	return hostsprint
}
