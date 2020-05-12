import _ from "lodash";
import React, { FC } from "react";

import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  Grid,
  Typography,
} from "@material-ui/core";
import { makeStyles } from "@material-ui/core/styles";
import SelectField from "../../../SelectField";

import { useMutation, useQuery } from "@apollo/react-hooks";
import { UPDATE_BILLING_INFO } from "../../../../apollo/queries/billinginfo";
import { QUERY_BILLING_METHODS } from "../../../../apollo/queries/billingmethod";
import { BillingInfo_billingInfo } from "../../../../apollo/types/BillingInfo";
import { BillingMethods, BillingMethodsVariables } from "../../../../apollo/types/BillingMethods";
import { OrganizationByName_organizationByName_PrivateOrganization } from "../../../../apollo/types/OrganizationByName";
import { UpdateBillingInfo, UpdateBillingInfoVariables } from "../../../../apollo/types/UpdateBillingInfo";
import billing from "../../../../lib/billing";

const useStyles = makeStyles((theme) => ({
  title: {
    marginTop: theme.spacing(4),
    marginBottom: theme.spacing(2),
  },
  button: {
    marginTop: theme.spacing(3),
    marginBotton: theme.spacing(2),
    marginRight: theme.spacing(3),
  },
  selectField: {
    marginTop: theme.spacing(1),
    minWidth: 300,
  },
  option: {
    fontSize: 15,
    "& > span": {
      marginRight: 10,
      fontSize: 18,
    },
    color: "white",
  },
  errorMsg: {
    marginTop: theme.spacing(3),
  },
}));

interface Props {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
  closeDialogue: (confirmationMessage: string) => void;
  billingInfo: BillingInfo_billingInfo;
}

const ChangeBillingMethod: FC<Props> = ({ organization, closeDialogue, billingInfo }) => {
  const classes = useStyles();
  const [errorDialogue, setErrorDialogue] = React.useState(false);
  const [error, setError] = React.useState("");
  const [billingMethodID, setBillingMethodID] = React.useState("");

  const { loading, error: queryError, data } = useQuery<BillingMethods, BillingMethodsVariables>(
    QUERY_BILLING_METHODS,
    {
      variables: { organizationID: organization.organizationID },
    }
  );

  const [updateBillingInfo, { error: mutError }] = useMutation<UpdateBillingInfo, UpdateBillingInfoVariables>(
    UPDATE_BILLING_INFO,
    {
      onCompleted: (data) => {
        if (data) {
          closeDialogue("Your billing method has been changed.");
        }
      },
    }
  );

  if (queryError || !data) {
    return <p>Error: {JSON.stringify(queryError)}</p>;
  }

  let billingMethodOptions: any[] = [];
  if (data.billingMethods && data.billingMethods.length > 0) {
    billingMethodOptions = data.billingMethods.map((billingMethod) => {
      if (billingMethod.paymentsDriver === billing.STRIPECARD_DRIVER) {
        const payload = JSON.parse(billingMethod.driverPayload);
        return {
          label: payload.brand.charAt(0).toUpperCase() + payload.brand.slice(1) + " xxxx-xxxx-xxxx-" + payload.last4,
          value: billingMethod.billingMethodID,
        };
      } else if (billingMethod.paymentsDriver === billing.STRIPEWIRE_DRIVER) {
        return {
          label: "Wire payment",
          value: billingMethod.billingMethodID,
        };
      } else if (billingMethod.paymentsDriver === billing.ANARCHISM_DRIVER) {
        return {
          label: "Anarchy!",
          value: billingMethod.billingMethodID,
        };
      } else {
        return { label: "Unrecognized payment driver.", value: "TODO" };
      }
    });
  }

  return (
    <>
      <Grid container direction="column">
        <Grid item>
          <SelectField
            id="billing_method"
            label="Billing method"
            value={billingMethodID}
            options={billingMethodOptions}
            onChange={({ target }) => setBillingMethodID( target.value as string )}
            controlClass={classes.selectField}
          />
        </Grid>
      </Grid>
      <Grid container spacing={2} className={classes.button}>
        <Grid item>
          <Button
            color="primary"
            autoFocus
            onClick={() => {
              closeDialogue("");
            }}
          >
            Cancel
          </Button>
        </Grid>
        <Grid item>
          <Button
            color="primary"
            variant="contained"
            autoFocus
            onClick={() => {
              if (billingMethodID) {
                updateBillingInfo({
                  variables: {
                    organizationID: organization.organizationID,
                    billingMethodID,
                    billingPlanID: billingInfo.billingPlan.billingPlanID,
                    country: billingInfo.country,
                  },
                });
              } else {
                setError("Please select your billing method.");
                setErrorDialogue(true);
              }
            }}
          >
            Change Billing Method
          </Button>
          {mutError && (
            <DialogContent>
              <Typography variant="body1" color="error" className={classes.errorMsg}>
                {mutError.message.replace("GraphQL error: ", "")}
              </Typography>
            </DialogContent>
          )}
          <Dialog open={errorDialogue} aria-describedby="alert-dialog-description">
            <DialogContent>
              {errorDialogue && (
                <Typography variant="body1" color="error">
                  {error}
                </Typography>
              )}
            </DialogContent>
            <DialogActions>
              <Button onClick={() => setErrorDialogue(false)} color="primary" autoFocus>
                Ok
              </Button>
            </DialogActions>
          </Dialog>
        </Grid>
      </Grid>
    </>
  );
};

export default ChangeBillingMethod;
