package main

import "fmt"

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
