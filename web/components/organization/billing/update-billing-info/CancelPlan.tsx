import _ from "lodash";
import React, { FC } from "react";

import {
  Button,
  DialogContent,
  DialogContentText,
  Grid,
  Typography,
} from "@material-ui/core";
import { makeStyles } from "@material-ui/core/styles";

import { useMutation, useQuery } from "@apollo/react-hooks";
import { UPDATE_BILLING_INFO } from "../../../../apollo/queries/billinginfo";
import { QUERY_BILLING_PLANS } from "../../../../apollo/queries/billingplan";
import { QUERY_ORGANIZATION } from "../../../../apollo/queries/organization";
import { BillingInfo_billingInfo } from "../../../../apollo/types/BillingInfo";
import { BillingPlans } from "../../../../apollo/types/BillingPlans";
import { OrganizationByName_organizationByName_PrivateOrganization } from "../../../../apollo/types/OrganizationByName";
import { UpdateBillingInfo, UpdateBillingInfoVariables } from "../../../../apollo/types/UpdateBillingInfo";

const useStyles = makeStyles((theme) => ({
  button: {
    marginTop: theme.spacing(3),
    marginBotton: theme.spacing(2),
    marginRight: theme.spacing(3),
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

const CancelPlan: FC<Props> = ({ organization, closeDialogue, billingInfo }) => {
  const classes = useStyles();

  const { loading, error: queryError, data: data1 } = useQuery<BillingPlans>(QUERY_BILLING_PLANS);

  const [updateBillingInfo, { error: mutError }] = useMutation<UpdateBillingInfo, UpdateBillingInfoVariables>(
    UPDATE_BILLING_INFO,
    {
      onCompleted: (data) => {
        if (data) {
          closeDialogue("Your plan has been canceled.");
        }
      },
      refetchQueries: [{ query: QUERY_ORGANIZATION, variables: { name: organization.name } }],
      awaitRefetchQueries: true,
    }
  );

  if (queryError || !data1) {
    return <p>Error: {JSON.stringify(queryError)}</p>;
  }

  const freePlan = data1.billingPlans.filter((billingPlan) => billingPlan.default)[0];

  return (
    <>
      <DialogContentText id="alert-dialog-description">
        Upon canceling your plan, your usage will be assessed and you will be charged for any applicable overage
        fees for the current billing period.
      </DialogContentText>
      <Grid container spacing={2} className={classes.button}>
        <Grid item>
          <Button color="primary" autoFocus onClick={() => closeDialogue("")}>
            No, go back
          </Button>
        </Grid>
        <Grid item>
          <Button
            color="primary"
            autoFocus
            onClick={() => {
              updateBillingInfo({
                variables: {
                  organizationID: organization.organizationID,
                  billingPlanID: freePlan.billingPlanID,
                  country: billingInfo.country,
                },
              });
            }}
          >
            Yes, I'm sure
          </Button>
          {mutError && (
            <DialogContent>
              <Typography variant="body1" color="error" className={classes.errorMsg}>
                {mutError.message.replace("GraphQL error: ", "")}
              </Typography>
            </DialogContent>
          )}
        </Grid>
      </Grid>
    </>
  );
};

export default CancelPlan;
