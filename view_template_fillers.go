package main

import (
	"html/template"
)

// bhapFiller fills the BHAP viewer page template.
type bhapFiller struct {
	ID            int
	PaddedID      string
	Title         string
	LastModified  string
	Author        string
	Status        status
	CreatedDate   string
	VotingOptions votingOptions
	HTMLContent   template.HTML
}

// votingOptions configures what voting options the user has in the BHAP
// screen.
type votingOptions struct {
	ShowReadyForDiscussion bool
	ShowAccept             bool
	ShowReject             bool
	ShowWithdraw           bool
	ShowReplace            bool
}

// list Filler fills the BHAP list page template.
type listFiller struct {
	LoggedIn bool
	Email    string
	Items    []listItemFiller
}

// listItemFiller fills a single BHAP in the BHAP list page template.
type listItemFiller struct {
	ID    int
	Title string
}

// newUserFiller fills the new user sign-up page template.
type newUserFiller struct {
	InvitationUID string
	Email         string
}

// bhapsToListItemFillers converts a BHAP to a list item filler, so it can be
// displayed in the BHAP list template.
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
