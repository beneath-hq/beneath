import {useQuery } from "@apollo/client";
import _ from "lodash";
import { Button, Table, TableBody, TableCell, TableHead, TableRow } from "@material-ui/core";
import React, { FC } from "react";

import { OrganizationByName_organizationByName_PrivateOrganization } from "apollo/types/OrganizationByName";
import ContentContainer, { CallToAction } from "components/ContentContainer";
import VSpace from "components/VSpace";
import { QUERY_BILLING_METHODS } from "ee/apollo/queries/billingMethod";
import { BillingMethods, BillingMethodsVariables } from "ee/apollo/types/BillingMethods";
import {ANARCHISM_DRIVER, STRIPECARD_DRIVER, STRIPEWIRE_DRIVER} from "ee/lib/billing";
import { BillingInfo_billingInfo } from "ee/apollo/types/BillingInfo";

export interface BillingMethodsProps {
  organization: OrganizationByName_organizationByName_PrivateOrganization;
  billingInfo: BillingInfo_billingInfo;
  addCard: (value: boolean) => void;
}

const ViewBillingMethods: FC<BillingMethodsProps> = ({ organization, billingInfo, addCard }) => {
  const { loading, error, data } = useQuery<BillingMethods, BillingMethodsVariables>(QUERY_BILLING_METHODS, {
    context: { ee: true },
    variables: { organizationID: organization.organizationID },
  });

  if (error) {
    return <p>Error: {JSON.stringify(error)}</p>;
  }

  if (!data) {
    return <></>;
  }

  if (data.billingMethods.length === 0) {
    const cta: CallToAction = {
      message: `You have no billing methods on file`,
      buttons: [{ label: "Add a credit card", onClick: () => addCard(true) }]
    };
    return (
      <ContentContainer callToAction={cta} />
    );
  }

  const formatDriver = (driver: string) => {
    if (driver === STRIPECARD_DRIVER) return "Card";
    if (driver === STRIPEWIRE_DRIVER) return "Wire";
    if (driver === ANARCHISM_DRIVER) return "Anarchism";
  };

  const formatDetails = (driver: string, driverPayload: string) => {
    if (driver === STRIPECARD_DRIVER) {
      const payload = JSON.parse(driverPayload);
      const brand = payload.brand.toString();
      const last4 = payload.last4.toString();
      const expMonth = payload.expMonth.toString();
      const expYear = payload.expYear.toString();
      return `${brand.toUpperCase()} ${last4}, Exp: ${expMonth}/${expYear.substring(2,4)}`;
    }
    if (driver === STRIPEWIRE_DRIVER) {
      const payload = JSON.parse(driverPayload);
      return `Invoices sent to ${payload.email_address}`;
    }
  };

  return (
    <>
      <Table>
        <TableHead>
          <TableRow>
            <TableCell>Type</TableCell>
            <TableCell>Details</TableCell>
            <TableCell>Active</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {data.billingMethods.map((billingMethod) => (
            <TableRow
              key={billingMethod.billingMethodID}
            >
              <TableCell>{formatDriver(billingMethod.paymentsDriver)}</TableCell>
              <TableCell>{formatDetails(billingMethod.paymentsDriver, billingMethod.driverPayload)}</TableCell>
              <TableCell>{billingMethod.billingMethodID === billingInfo.billingMethod?.billingMethodID ? "Yes" : "No"}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
      <VSpace units={2} />
      <Button onClick={() => addCard(true)}>Add card</Button>
    </>
  );
};

export default ViewBillingMethods;