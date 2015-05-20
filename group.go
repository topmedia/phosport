package main

import (
	"encoding/json"
	"fmt"
	"log"
)

type Group struct {
	Data  GroupData `json:"data"`
	Ref   string    `json:"ref"`
	Class string    `json:"class"`
	Type  string    `json:"type"`
}

type GroupData struct {
	Comment string   `json:"comment"`
	Ref     string   `json:"ref"`
	Name    string   `json:"name"`
	Members []string `json:"members"`
}

func ExportGroups(class string) []GroupData {
	objs := ConfdCommand("get_objects_filtered", fmt.Sprintf(`$_->{type} eq "group" && $_->{class} eq "%s"`, class))

	var groups []Group
	err := json.Unmarshal(ToJSON(objs), &groups)
	if err != nil {
		log.Fatalf("Error parsing groups into JSON: %s", err)
	}

	groupsprint := make([]GroupData, len(groups))

	for _, group := range groups {
		group.Data.Ref = group.Ref
		groupsprint = append(groupsprint, group.Data)
	}
	return groupsprint
}
