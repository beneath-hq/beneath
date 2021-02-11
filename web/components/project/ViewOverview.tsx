import React, { FC } from "react";

import {
  ProjectByOrganizationAndName_projectByOrganizationAndName,
  ProjectByOrganizationAndName_projectByOrganizationAndName_streams,
  ProjectByOrganizationAndName_projectByOrganizationAndName_services,
} from "apollo/types/ProjectByOrganizationAndName";
import clsx from "clsx";
import ContentContainer, { CallToAction } from "components/ContentContainer";
import { Table, TableBody, TableCell, TableHead, TableLinkRow, TableRow } from "components/Tables";
import { toURLName } from "lib/names";
import { Chip, Grid, Hidden, makeStyles } from "@material-ui/core";
import { MetaChip, StreamUsageChip } from "components/stream/chips";

const useStyles = makeStyles((theme) => ({
  nameCell: {
    whiteSpace: "nowrap",
  },
  descriptionCell: {
    maxWidth: "400px",
    whiteSpace: "nowrap",
    overflow: "hidden",
    textOverflow: "ellipsis",
  },
  pointer: {
    cursor: "pointer",
  },
  streamChip: {
    backgroundColor: theme.palette.primary.dark,
  },
  serviceChip: {
    backgroundColor: theme.palette.purple.main,
  },
}));

export interface ViewOverviewProps {
  project: ProjectByOrganizationAndName_projectByOrganizationAndName;
}

type Stream = ProjectByOrganizationAndName_projectByOrganizationAndName_streams;
type Service = ProjectByOrganizationAndName_projectByOrganizationAndName_services;
type Resource = Stream | Service;

const ViewOverview: FC<ViewOverviewProps> = ({ project }) => {
  const classes = useStyles();

  const resources: Resource[] = [...project.streams, ...project.services];

  let cta: CallToAction | undefined;
  if (!resources.length) {
    cta = {
      message: (
        <>
          We didn't find any items in{" "}
          <strong>
            {toURLName(project.organization.name)}/{toURLName(project.name)}
          </strong>
        </>
      ),
    };
    if (project.permissions.create) {
      cta.buttons = [
        {
          label: "Create stream",
          href: `/-/create/stream?organization=${project.organization.name}&project=${project.name}`,
          as: "/-/create/stream",
        },
      ];
    }
  }

  const makeHref = (resource: Resource) => {
    if (resource.__typename === "Stream") {
      return `/stream?organization_name=${toURLName(project.organization.name)}&project_name=${toURLName(
        project.name
      )}&stream_name=${toURLName(resource.name)}`;
    } else if (resource.__typename === "Service") {
      return `/service?organization_name=${toURLName(project.organization.name)}&project_name=${toURLName(
        project.name
      )}&service_name=${toURLName(resource.name)}`;
    }
    return "";
  };

  const makeAs = (resource: Resource) => {
    if (resource.__typename === "Stream") {
      return `/${toURLName(project.organization.name)}/${toURLName(project.name)}/stream:${toURLName(resource.name)}`;
    } else if (resource.__typename === "Service") {
      return `/${toURLName(project.organization.name)}/${toURLName(project.name)}/service:${toURLName(resource.name)}`;
    }
    return "";
  };

  return (
    <ContentContainer paper callToAction={cta}>
      <Table textSize="medium">
        <TableHead>
          <TableRow>
            <TableCell></TableCell>
            <TableCell>Name</TableCell>
            <Hidden smDown>
              <TableCell expand>Description</TableCell>
            </Hidden>
            <Hidden xsDown>
              <TableCell></TableCell>
            </Hidden>
          </TableRow>
        </TableHead>
        <TableBody>
          {Array.from(resources)
            .sort((a, b) => a.name.localeCompare(b.name))
            .map((resource) => (
              <TableLinkRow key={resource.name + resource.createdOn} href={makeHref(resource)} as={makeAs(resource)}>
                <TypeTableCell resource={resource} />
                <TableCell className={classes.nameCell}>{toURLName(resource.name)}</TableCell>
                <Hidden smDown>
                  <TableCell className={classes.descriptionCell}>{resource.description}</TableCell>
                </Hidden>
                <Hidden xsDown>
                  <TableCell align="right">
                    {resource.__typename === "Stream" && (
                      <Grid container wrap="nowrap" justify="flex-end" spacing={1}>
                        {resource.meta && (
                          <Grid item>
                            <MetaChip />
                          </Grid>
                        )}
                        <Grid item>
                          <StreamUsageChip stream={{ ...resource, project }} />
                        </Grid>
                      </Grid>
                    )}
                  </TableCell>
                </Hidden>
              </TableLinkRow>
            ))}
        </TableBody>
      </Table>
    </ContentContainer>
  );
};

export default ViewOverview;

const TypeTableCell: FC<{ resource: Resource }> = ({ resource }) => {
  const classes = useStyles();
  return (
    <TableCell>
      <Chip
        label={<strong>{resource.__typename}</strong>}
        className={clsx(
          classes.pointer,
          resource.__typename === "Stream" && classes.streamChip,
          resource.__typename === "Service" && classes.serviceChip
        )}
      />
    </TableCell>
  );
};
