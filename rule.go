package main

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

type RulePrint struct {
	Sources      []string `json:"sources"`
	Destinations []string `json:"destinations"`
	Services     []string `json:"services"`
}
