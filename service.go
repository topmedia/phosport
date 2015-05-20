package main

import (
	"encoding/json"
	"fmt"
	"log"
)

type Service struct {
	Data  ServiceData `json:"data"`
	Class string      `json:"class"`
	Ref   string      `json:"ref"`
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
	ServiceDataPrint
	Members []string `json:"members"`
}

func (sd *ServiceData) MarshalJSON() ([]byte, error) {
	sdp := ServiceDataPrint{
		DstHigh:  sd.DstHigh,
		DstLow:   sd.DstLow,
		Comment:  sd.Comment,
		Name:     sd.Name,
		SrcHigh:  sd.SrcHigh,
		SrcLow:   sd.SrcLow,
		Ref:      sd.Ref,
		Protocol: sd.Protocol,
	}

	return json.Marshal(sdp)
}

type ServiceDataPrint struct {
	DstHigh  int    `json:"dst_high"`
	DstLow   int    `json:"dst_low"`
	Comment  string `json:"comment"`
	Name     string `json:"name"`
	SrcHigh  int    `json:"src_high"`
	SrcLow   int    `json:"src_low"`
	Ref      string `json:"ref"`
	Protocol string `json:"protocol"`
}

func ExportServices() []ServiceData {
	objs := ConfdCommand("get_objects_filtered", `$_->{class} eq "service" && $_->{type} =~ /(tcp|udp|icmp)/`)

	var services []Service
	err := json.Unmarshal(ToJSON(objs), &services)
	if err != nil {
		log.Fatalf("Error parsing services into JSON: %s", err)
	}

	servicesprint := make([]ServiceData, len(services))

	for _, svc := range services {
		svc.Data.Ref = svc.Ref
		svc.Data.Protocol = svc.Type
		servicesprint = append(servicesprint, svc.Data)
	}
	return servicesprint
}
