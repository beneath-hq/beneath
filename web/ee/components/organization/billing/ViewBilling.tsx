import { useQuery } from "@apollo/client";
import _ from "lodash";
import { Container, Dialog, DialogContent, DialogTitle, Link } from "@material-ui/core";
import { Alert } from "@material-ui/lab";
import { makeStyles } from "@material-ui/core/styles";
import dynamic from "next/dynamic";
import React, { FC } from "react";

import { OrganizationByName_organizationByName_PrivateOrganization } from "apollo/types/OrganizationByName";
import VSpace from "components/VSpace";
import { BillingInfo, BillingInfoVariables } from "ee/apollo/types/BillingInfo";
import { QUERY_BILLING_INFO } from "ee/apollo/queries/billingInfo";
import ViewTaxInfo from "./tax-info/ViewTaxInfo";
import EditTaxInfo from "./tax-info/EditTaxInfo";
import ViewBillingPlan from "./billing-plan/ViewBillingPlan";
import ViewBillingMethods from "./billing-method/ViewBillingMethods";

const useStyles = makeStyles((theme) => ({
  container: {
    padding: "0px",
  },
}));

export interface ViewBillingProps {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
}

const ViewBilling: FC<ViewBillingProps> = ({ organization }) => {
  const classes = useStyles();
  const [addCardDialog, setAddCardDialog] = React.useState(false);
  const [editTaxInfoDialog, setEditTaxInfoDialog] = React.useState(false);
  const DynamicCardForm = dynamic(() => import("./billing-method/CardForm"));

  const { loading, error, data } = useQuery<BillingInfo, BillingInfoVariables>(QUERY_BILLING_INFO, {
    context: { ee: true },
    variables: {
      organizationID: organization.organizationID,
    },
  });

  if (error || !data) {
    return <p>Error: {JSON.stringify(error)}</p>;
  }

  // alert when you're viewing billing for your personal organization but your main billing is handled by another org
  const specialCase =
    organization.personalUser && organization.personalUser.billingOrganizationID !== organization.organizationID;

  return (
    <React.Fragment>
      <Container maxWidth="md" className={classes.container}>
        {specialCase && (
          <>
            <Alert severity="info">
              Note that you are a member of the {organization.personalUser?.billingOrganization.displayName}{" "}
              organization, and so {organization.personalUser?.billingOrganization.displayName} is billed for your
              Beneath usage. However, you are individually billed for the activity of any services that you have not
              transferred to {organization.personalUser?.billingOrganization.displayName}.
              <br />
              <br /> If you have no active services managed by your user, then make sure to cancel any active billing
              plan on this page. Then you won't incur any unneccessary charges.
              <br />
              <br />
              If you have active services that you would like{" "}
              {organization.personalUser?.billingOrganization.displayName} to pay for, you need to transfer those
              services from your personal user to the {organization.personalUser?.billingOrganization.displayName}{" "}
              organization. See the docs <Link href="https://about.beneath.dev/docs/core-resources/services">here</Link>
              .
              <br />
              <br /> If you have active services that you would like to pay for yourself, and not assign to your
              organization, then you must ensure the billing information on this page covers you. If the services' usage
              exceeds the quotas of the Free tier, then you should upgrade your personal billing plan on this page.
            </Alert>
            <VSpace units={1} />
          </>
        )}

        <Alert severity="info">
          You can find detailed information about our billing plans{" "}
          <Link href="https://about.beneath.dev/pricing/">here</Link>.
        </Alert>

        <VSpace units={3} />

        <ViewBillingPlan organization={organization} />

        <VSpace units={3} />

        <ViewBillingMethods organization={organization} billingInfo={data.billingInfo} addCard={setAddCardDialog} />
        <Dialog open={addCardDialog} fullWidth={true} maxWidth={"sm"}>
          <DialogTitle id="alert-dialog-title">{"Add a card"}</DialogTitle>
          <DialogContent>
            <DynamicCardForm organization={organization} openDialogFn={setAddCardDialog} />
          </DialogContent>
        </Dialog>

        <VSpace units={3} />

        <ViewTaxInfo organization={organization} onEdit={() => setEditTaxInfoDialog(true)} />
        <Dialog open={editTaxInfoDialog} onBackdropClick={() => setEditTaxInfoDialog(false)} fullWidth maxWidth="sm">
          <DialogTitle>Edit tax info</DialogTitle>
          <DialogContent>
            <EditTaxInfo
              organization={organization}
              billingInfo={data.billingInfo}
              editTaxInfo={setEditTaxInfoDialog}
            />
          </DialogContent>
        </Dialog>
      </Container>
    </React.Fragment>
  );
};

export default ViewBilling;
