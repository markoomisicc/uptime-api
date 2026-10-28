package api

import "fmt"

func verificationEmail(language, firstName, link string) (subject, body string) {
	switch language {
	case "de":
		return "Bestätige deine E-Mail-Adresse", fmt.Sprintf(
			"Hallo %s,\n\nbitte bestätige deine E-Mail-Adresse, indem du auf den folgenden Link klickst:\n\n%s\n\nDer Link ist 24 Stunden gültig.\n\nFalls du dich nicht registriert hast, kannst du diese E-Mail ignorieren.\n",
			firstName, link)
	default:
		return "Confirm your email address", fmt.Sprintf(
			"Hi %s,\n\nplease confirm your email address by clicking the link below:\n\n%s\n\nThe link is valid for 24 hours.\n\nIf you didn't sign up, you can ignore this email.\n",
			firstName, link)
	}
}
