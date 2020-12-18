import { Box, Container } from "@material-ui/core";
import { FC } from "react";

import { ServiceByOrganizationProjectAndName_serviceByOrganizationProjectAndName } from "../../apollo/types/ServiceByOrganizationProjectAndName";
import ListSecrets from "./access/ListSecrets";
import ListPermissions from "./access/ListPermissions";

export interface Props {
  service: ServiceByOrganizationProjectAndName_serviceByOrganizationProjectAndName;
}

const ViewAccess: FC<Props> = ({ service }) => {
  return (
    <>
      <Container maxWidth="md">
        <ListPermissions serviceID={service.serviceID} />
        <Box m={4} />
        <ListSecrets serviceID={service.serviceID} />
      </Container>
    </>
  );
};

export default ViewAccess;
