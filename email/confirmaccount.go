package email

import "github.com/t0nyandre/go-rest-boilerplate/extras"

// ConfirmAccountEmail for sending out confirmation emails when a user creates an account
func ConfirmAccountEmail(from string, to []string, data interface{}) {
	email := Email{
		From:    from,
		To:      to,
		Data:    data,
		Subject: "Just one more step to go ...",
		Text:    "Thank you so much for signing up to my site! All you need now is to confirm your account!",
		Images:  []string{"logo.png"},
	}

	email.sendEmail(string(extras.ConfirmAccount))
}
