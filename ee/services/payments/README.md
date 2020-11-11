# Payments

The payments service handles invoicing and charging bills. Contrast with the `billing` service, which is responsible for *what to bill from who*.
 
## Stripe
To get a grip on how the Stripe Go library works, check out the [Usage](https://github.com/stripe/stripe-go#usage) section.

### Development
To test Stripe's webhooks in your local development environment, execute this command in a terminal: `stripe listen --forward-to localhost:4000/ee/billing/stripecard/webhook`.

### Production

In the Stripe Developer Dashboard, we register the production webhook endpoint: https://control.beneath.dev/ee/billing/stripecard/webhook. And we tell Stripe which events we would like to receive. Need to make sure we sign-up for all events that we handle in our `handleStripeWebhook()` function.

## How to change a customer's payment method

- To allow a customer to pay by wire, submit an http request to http://control.beneath.dev/ee/billing/stripewire/initialize_customer with these parameters: organization_id, email_address
