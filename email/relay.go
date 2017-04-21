package email

type EmailRelay struct {
	// Callback function that should perform the email sending process
	// password. Must return an error on failure. Required.
	Send func(email *Email) error
}
