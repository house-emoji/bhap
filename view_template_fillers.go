package main

import (
	"html/template"
)

type bhapFiller struct {
	PaddedID     string
	Title        string
	LastModified string
	Author       string
	Status       status
	CreatedDate  string
	HTMLContent  template.HTML
}

type listItemFiller struct {
	ID    int
	Title string
}

type newUserFiller struct {
	InvitationUID string
	Email         string
}

func bhapsToListItemFillers(bhaps []bhap) []listItemFiller {
	filler := make([]listItemFiller, 0)
	for _, val := range bhaps {
		filler = append(filler, listItemFiller{
			ID:    val.ID,
			Title: val.Title,
		})
	}
	return filler
}
