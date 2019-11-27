import React, { FC } from 'react';
import { useMutation } from "@apollo/react-hooks";
import { injectStripe, ReactStripeElements } from 'react-stripe-elements';
import { CardElement } from 'react-stripe-elements';
import { CREATE_STRIPE_SETUP_INTENT } from "../../apollo/queries/organization";
import { CreateStripeSetupIntent, CreateStripeSetupIntentVariables } from "../../apollo/types/CreateStripeSetupIntent";
import { OrganizationByName_organizationByName } from "../../apollo/types/OrganizationByName";
import { makeStyles, TextField, Typography } from "@material-ui/core";

interface CheckoutParams {
  organization: OrganizationByName_organizationByName
}

class CheckoutFormWrapper extends React.Component<ReactStripeElements.InjectedStripeProps & CheckoutParams, CheckoutParams> {
  constructor(props: ReactStripeElements.InjectedStripeProps & CheckoutParams) {
    super(props);
    this.state = { organization: props.organization };
  }

  render() {
    return <CheckoutForm stripe={this.props.stripe} organization={this.state.organization} />
  }
}

export default injectStripe(CheckoutFormWrapper);


const useStyles = makeStyles((theme) => ({
  noDataCaption: {
    color: theme.palette.text.secondary,
  },
}));

interface Props {
  stripe: ReactStripeElements.StripeProps | undefined;
  organization: OrganizationByName_organizationByName;
}

const CheckoutForm: FC<Props> = ({ stripe, organization }) => {
  const classes = useStyles();
  const PRO_BILLING_PLAN_ID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa";
  
  const [errorMsg, setErrorMsg] = React.useState("")
  const [status, setStatus] = React.useState("")
  const [customerData, setCustomerData] = React.useState({
    payment_method_data: {
      billing_details: {
        address: {
          city: "",
          country: "",
          line1: "",
          line2: "",
          postal_code: "",
          state: ""
        },
        email: "",
        name: "",
        phone: ""
      }
    }
  })

  const [createStripeSetupIntent, { loading, error }] = useMutation<CreateStripeSetupIntent, CreateStripeSetupIntentVariables>(CREATE_STRIPE_SETUP_INTENT, {
    onCompleted: (data) => {
      if (!stripe) {
        return;
      }

      // handleCardSetup automatically pulls credit card info from the Card element
      // TODO from Stripe Docs: Note that stripe.handleCardSetup may take several seconds to complete. During that time, you should disable your form from being resubmitted and show a waiting indicator like a spinner. If you receive an error result, you should be sure to show that error to the customer, re-enable the form, and hide the waiting indicator.
      // TODO from Stripe Docs: Additionally, stripe.handleCardSetup may trigger a 3D Secure authentication challenge.This will be shown in a modal dialog and may be confusing for customers using assistive technologies like screen readers.You should make your form accessible by ensuring that success or error messages are clearly read out after this method completes
      stripe.handleCardSetup(data.createStripeSetupIntent, customerData).then(
        result => {
          if (result.error) {
            if (result.error.message) {
              setErrorMsg(result.error.message)
            }
          }
          if (result.setupIntent) {
            setStatus(result.setupIntent.status)
          }
        })
      return
    }
  });

  if (!stripe) {
    return <p>Loading or Error</p>;
  }

  const handleSubmit = (ev: any) => {
    // We don't want to let default form submission happen here, which would refresh the page.
    ev.preventDefault();
    
    setErrorMsg("")
    setStatus("")
    // TODO: get customerData from Form below
    setCustomerData(
      {
        payment_method_data: {
          billing_details: {
            address: {
              city: "Boston",
              country: "US",
              line1: "74 Stone Rd",
              line2: "none",
              postal_code: "02478",
              state: "MA"
            },
            email: "ericpgreen2@gmail.com",
            name: "Eric Green",
            phone: "6177101732"
          }
        }
      }
    )

    createStripeSetupIntent({ variables: { organizationID: organization.organizationID, billingPlanID: PRO_BILLING_PLAN_ID } });

    return
  };

  return (
    <form onSubmit={handleSubmit}>
      {/* <TextField
        id="firstname"
        label="Firstname"
        margin="normal"
        fullWidth
        required
      />
      <TextField
        id="lastname"
        label="Lastname"
        margin="normal"
        fullWidth
        required
      />
      <TextField
        id="address_line1"
        label="Address_line1"
        margin="normal"
        fullWidth
        required
      />
      <TextField
        id="address_line2"
        label="Address_line2"
        margin="normal"
        fullWidth
      />
      <TextField
        id="address_city"
        label="Address_city"
        margin="normal"
        fullWidth
        required
      />
      <TextField
        id="address_state"
        label="Address_state"
        margin="normal"
        fullWidth
        required
      />
      <TextField
        id="address_zip"
        label="Address_zip"
        margin="normal"
        fullWidth
        required
      />
      <TextField
        id="address_country"
        label="Address_country"
        margin="normal"
        fullWidth
        required
      /> */}
      <CardElement style={{ base: { fontSize: '18px', color: '#FFFFFF' } }} />
      <button>Submit</button>
      {errorMsg !== "" && (
        <Typography variant="body1" color="error">
          { errorMsg }
        </Typography>
      )}
      {status.length > 0 && (
        <Typography variant="body1" color="error">
          { status }
        </Typography>
      )}
    </form>
  );
};
