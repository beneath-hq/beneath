package sendgrid

import (
	"fmt"

	"gitlab.com/beneath-hq/beneath/pkg/envutil"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

const fromBeneathName = "Eric Green"
const fromBeneathEmail = "eric@beneath.dev"
const welcomeEmailTemplateID = "d-efee7240ee394855b86ac0705c3dd4f5"

type configSpecification struct {
	SendGridAPIKey string `envconfig:"BI_SENDGRID_API_KEY" required:"true"`
}

// SendWelcomeEmail sends the welcome email via SendGrid
func SendWelcomeEmail(emailAddress string) error {
	var config configSpecification
	envutil.LoadConfig("beneath", &config)

	m := mail.NewV3Mail()
	e := mail.NewEmail(fromBeneathName, fromBeneathEmail)
	m.SetFrom(e)
	m.SetTemplateID(welcomeEmailTemplateID)

	p := mail.NewPersonalization()
	tos := []*mail.Email{
		mail.NewEmail("", emailAddress),
	}
	p.AddTos(tos...)
	m.AddPersonalizations(p)

	request := sendgrid.GetRequest(config.SendGridAPIKey, "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	var Body = mail.GetRequestBody(m)
	request.Body = Body
	response, err := sendgrid.API(request)

	if err != nil {
		return fmt.Errorf("Error sending email: %v\\n", err)
	} else if response.StatusCode >= 400 {
		return fmt.Errorf("%d error sending email: %v\\n", response.StatusCode, response.Body)
	}

	// done
	return nil
}
