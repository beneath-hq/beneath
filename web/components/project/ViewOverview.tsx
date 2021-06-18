import React, { FC } from "react";

import {
  ProjectByOrganizationAndName_projectByOrganizationAndName,
  ProjectByOrganizationAndName_projectByOrganizationAndName_tables,
  ProjectByOrganizationAndName_projectByOrganizationAndName_services,
} from "apollo/types/ProjectByOrganizationAndName";
import clsx from "clsx";
import ContentContainer, { CallToAction } from "components/ContentContainer";
import { UITable, UITableBody, UITableCell, UITableHead, UITableLinkRow, UITableRow } from "components/UITables";
import { toURLName } from "lib/names";
import { Chip, Grid, Hidden, makeStyles } from "@material-ui/core";
import { MetaChip, TableUsageChip } from "components/table/chips";

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
  tableChip: {
    backgroundColor: theme.palette.primary.dark,
  },
  serviceChip: {
    backgroundColor: theme.palette.purple.main,
  },
}));

export interface ViewOverviewProps {
  project: ProjectByOrganizationAndName_projectByOrganizationAndName;
}

type Table = ProjectByOrganizationAndName_projectByOrganizationAndName_tables;
type Service = ProjectByOrganizationAndName_projectByOrganizationAndName_services;
type Resource = Table | Service;

const ViewOverview: FC<ViewOverviewProps> = ({ project }) => {
  const classes = useStyles();

  const resources: Resource[] = [...project.tables, ...project.services];

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
          label: "Create table",
          href: `/-/create/table?organization=${project.organization.name}&project=${project.name}`,
          as: "/-/create/table",
        },
      ];
    }
  }

  const makeHref = (resource: Resource) => {
    if (resource.__typename === "Table") {
      return `/table?organization_name=${toURLName(project.organization.name)}&project_name=${toURLName(
        project.name
      )}&table_name=${toURLName(resource.name)}`;
    } else if (resource.__typename === "Service") {
      return `/service?organization_name=${toURLName(project.organization.name)}&project_name=${toURLName(
        project.name
      )}&service_name=${toURLName(resource.name)}`;
    }
    return "";
  };

  const makeAs = (resource: Resource) => {
    if (resource.__typename === "Table") {
      return `/${toURLName(project.organization.name)}/${toURLName(project.name)}/table:${toURLName(resource.name)}`;
    } else if (resource.__typename === "Service") {
      return `/${toURLName(project.organization.name)}/${toURLName(project.name)}/service:${toURLName(resource.name)}`;
    }
    return "";
  };

  return (
    <ContentContainer paper callToAction={cta}>
      <UITable textSize="medium">
        <UITableHead>
          <UITableRow>
            <UITableCell></UITableCell>
            <UITableCell>Name</UITableCell>
            <Hidden smDown>
              <UITableCell expand>Description</UITableCell>
            </Hidden>
            <Hidden xsDown>
              <UITableCell></UITableCell>
            </Hidden>
          </UITableRow>
        </UITableHead>
        <UITableBody>
          {Array.from(resources)
            .sort((a, b) => a.name.localeCompare(b.name))
            .map((resource) => (
              <UITableLinkRow key={resource.name + resource.createdOn} href={makeHref(resource)} as={makeAs(resource)}>
                <TypeUITableCell resource={resource} />
                <UITableCell className={classes.nameCell}>{toURLName(resource.name)}</UITableCell>
                <Hidden smDown>
                  <UITableCell className={classes.descriptionCell}>{resource.description}</UITableCell>
                </Hidden>
                <Hidden xsDown>
                  <UITableCell align="right">
                    {resource.__typename === "Table" && (
                      <Grid container wrap="nowrap" justify="flex-end" spacing={1}>
                        {resource.meta && (
                          <Grid item>
                            <MetaChip />
                          </Grid>
                        )}
                        <Grid item>
                          <TableUsageChip table={{ ...resource, project }} notClickable />
                        </Grid>
                      </Grid>
                    )}
                  </UITableCell>
                </Hidden>
              </UITableLinkRow>
            ))}
        </UITableBody>
      </UITable>
    </ContentContainer>
  );
};

export default ViewOverview;

const TypeUITableCell: FC<{ resource: Resource }> = ({ resource }) => {
  const classes = useStyles();
  return (
    <UITableCell>
      <Chip
        label={<strong>{resource.__typename}</strong>}
        className={clsx(
          classes.pointer,
          resource.__typename === "Table" && classes.tableChip,
          resource.__typename === "Service" && classes.serviceChip
        )}
      />
    </UITableCell>
  );
};
