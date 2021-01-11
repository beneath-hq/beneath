import React, { FC } from "react";

import { ProjectByOrganizationAndName_projectByOrganizationAndName } from "apollo/types/ProjectByOrganizationAndName";
import Avatar from "components/Avatar";
import ContentContainer, { CallToAction } from "components/ContentContainer";
import { Table, TableBody, TableCell, TableHead, TableLinkRow, TableRow } from "components/Tables";
import { toURLName } from "lib/names";

export interface ViewServicesProps {
  project: ProjectByOrganizationAndName_projectByOrganizationAndName;
}

const ViewServices: FC<ViewServicesProps> = ({ project }) => {
  let cta: CallToAction | undefined;
  if (!project.services?.length) {
    cta = {
      message: `${project.displayName || project.name} doesn't have any services`,
    };
    if (project.permissions.create) {
      cta.buttons = [{ label: "Create service", href: "/-/create/service" }];
    }
  }

  return (
    <ContentContainer paper maxWidth={"md"} callToAction={cta}>
      <Table textSize="medium">
        <TableHead>
          <TableRow>
            <TableCell padding="checkbox"></TableCell>
            <TableCell>Name</TableCell>
            <TableCell>Description</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {project.services.map(({ serviceID, name, description }) => (
            <TableLinkRow
              key={serviceID}
              href={
                `/service?organization_name=${toURLName(project.organization.name)}` +
                `&project_name=${toURLName(project.name)}&service_name=${toURLName(name)}`
              }
              as={`/${toURLName(project.organization.name)}/${toURLName(project.name)}/-/services/${toURLName(name)}`}
            >
              <TableCell>
                <Avatar size="list" label={name} />
              </TableCell>
              <TableCell>{toURLName(name)}</TableCell>
              <TableCell>{description}</TableCell>
            </TableLinkRow>
          ))}
        </TableBody>
      </Table>
    </ContentContainer>
  );
};

export default ViewServices;
