# README

## Stripe

To test Stripe's webhooks in your local development environment, run in a terminal window: `stripe listen --forward-to localhost:4000/billing/stripecard/webhook`

## Changing payment plans

// TODO: need to polish these instructions and include all cases
To upgrade a customer to Enterprise, submit an http request to http://localhost:4000/billing/stripewire/initialize_customer with these parameters: organizationID, billingPlanID, emailAddress
To downgrade a customer to Free, submit an http request to http://localhost:4000/billing/anarchism/initialize_customer with these parameters: organizationID, billingPlanID
 