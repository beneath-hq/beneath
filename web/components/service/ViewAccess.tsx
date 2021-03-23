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
        <ListPermissions serviceID={service.serviceID} editable={service.project.permissions.create} />
        <Box m={4} />
        {service.project.permissions.create && <ListSecrets serviceID={service.serviceID} />}
      </Container>
    </>
  );
};

export default ViewAccess;
