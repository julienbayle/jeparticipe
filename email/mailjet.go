package email

import (
	"os"

	"github.com/mailjet/mailjet-apiv3-go"
)

// Sends mail using mailjet API
func SendWithMailjet(email *Email) error {

	// Get Mailjet keys from environnement
	publicKey := os.Getenv("MJ_APIKEY_PUBLIC")
	secretKey := os.Getenv("MJ_APIKEY_PRIVATE")

	mj := mailjet.NewMailjetClient(publicKey, secretKey)

	param := &mailjet.InfoSendMail{
		FromEmail: "no-reply@circuleo.fr",
		FromName:  "Circuleo - Evènement",
		Recipients: []mailjet.Recipient{
			mailjet.Recipient{
				Email: email.To,
			},
		},
		Subject:  email.Subject,
		TextPart: email.Body,
	}
	_, err := mj.SendMail(param)
	return err
}
