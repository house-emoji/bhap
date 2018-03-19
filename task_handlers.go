package main

import (
	"bytes"
	"net/http"
	"text/template"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/mail"
)

var invitationTemplate *template.Template

const InvitationSubject = "You have been invited to join the BHAP Consortium"

func init() {
	invitationTemplate = template.Must(template.ParseFiles("mail_templates/invitation.txt"))
}

func sendInvitations(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	unsents, keys, err := unsentInvitations(ctx)
	if err != nil {
		http.Error(w, "Could not get unsent invitations", 500)
		log.Errorf(ctx, "could not get unsent invitations: %v", err)
		return
	}

	log.Infof(ctx, "about to send %v invitations", len(unsents))

	failCount := 0

	for i, unsent := range unsents {
		var buf bytes.Buffer
		filler := invitationFiller{
			CreateAccountURL: "https://the-bhaps.appspot.com/invitation/" + unsent.UID,
		}
		if err := invitationTemplate.Execute(&buf, filler); err != nil {
			log.Errorf(ctx, "failed to execute email invitation template: %v", err)
			failCount++
			continue
		}

		message := mail.Message{
			Sender:  "BHAP Invitations <invitations@the-bhaps.appspot.com>",
			To:      []string{unsent.Email},
			Subject: InvitationSubject,
			Body:    buf.String(),
		}

		if err := mail.Send(ctx, &message); err != nil {
			log.Errorf(ctx, "failed to send mail to %v: %v",
				unsent.Email, err)
			failCount++
			continue
		}

		unsent.EmailSent = true
		if _, err := datastore.Put(ctx, keys[i], &unsent); err != nil {
			log.Errorf(ctx, "failed to save invitation: %v", err)
			failCount++
			continue
		}

		log.Infof(ctx, "sent an invitation to %v", unsent.Email)
	}

	if failCount > 0 {
		http.Error(w, "Failures while sending emails", 500)
		log.Infof(ctx, "failed the task because %v emails failed to send", failCount)
		return
	}
}
