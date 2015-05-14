package main

import "fmt"

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
