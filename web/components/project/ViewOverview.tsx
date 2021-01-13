import numbro from "numbro";
import React, { FC } from "react";

import { EntityKind } from "apollo/types/globalTypes";
import { ProjectByOrganizationAndName_projectByOrganizationAndName } from "apollo/types/ProjectByOrganizationAndName";
import ContentContainer, { CallToAction } from "components/ContentContainer";
import { useTotalUsage } from "components/usage/util";
import { Table, TableBody, TableCell, TableHead, TableLinkRow, TableRow } from "components/Tables";
import { toURLName } from "lib/names";
import { Chip, Grid, Hidden, makeStyles, Paper, Typography } from "@material-ui/core";
import clsx from "clsx";

// TODO (maybe): make a "Tag(text, color, size)" component to use instead of the wacky Paper styling for resourceType. 
// We have similar styling for the "tags" in the header of RecordsTable
const useStyles = makeStyles((theme) => ({
  resourceTypeCell: {
    maxWidth: "80px",
  },
  resourceTypePaper: {
    height: "20px",
    width: "60px",
    padding: theme.spacing(0.5),
  },
  resourceTypePaperText: {
    display: "table-cell",
    textAlign: "center",
    verticalAlign: "middle",
    lineHeight: "12px",
    fontWeight: "bold",
  },
  streamPaper: {
    backgroundColor: theme.palette.primary.dark,
  },
  servicePaper: {
    backgroundColor: theme.palette.purple.main,
  },
  nameCell: {
    whiteSpace: "nowrap",
  },
  descriptionCell: {
    maxWidth: "350px",
    whiteSpace: "nowrap",
    overflow: "hidden",
    textOverflow: "ellipsis",
  },
  chipsCell: {
    whiteSpace: "nowrap"
  },
  chip: {
    backgroundColor: theme.palette.secondary.dark,
    cursor: "pointer",
  },
}));

const intFormat = { thousandSeparated: true };
const bytesFormat: numbro.Format = { base: "decimal", mantissa: 1, optionalMantissa: true, output: "byte" };

export interface ViewOverviewProps {
  project: ProjectByOrganizationAndName_projectByOrganizationAndName;
}

interface resource {
  type: string;
  name: string;
  description: string | null;
  resourceID: string;
}

const ViewOverview: FC<ViewOverviewProps> = ({ project }) => {
  const classes = useStyles();
  const streams: resource[] = project.streams.map((stream) => {
    return { type: "stream", name: stream.name, description: stream.description, resourceID: stream.streamID }
  })
  const services: resource[] = project.services.map((service) => {
    return { type: "service", name: service.name, description: service.description, resourceID: service.serviceID }
  })
  const resources: resource[] = streams.concat(services);

  let cta: CallToAction | undefined;
  if (!resources.length) {  // TODO: might have to change this to === 0
    cta = {
      message: (
        <>
          We didn't find any resources in{" "}
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

  const makeHref = (resourceType: string, name: string) => {
    if (resourceType === "stream") {
      return `/stream?organization_name=${toURLName(project.organization.name)}&project_name=${toURLName(project.name)}&stream_name=${toURLName(name)}`
    } else { // resourceType === "service"
      return `/service?organization_name=${toURLName(project.organization.name)}&project_name=${toURLName(project.name)}&service_name=${toURLName(name)}`;
    }
  }

  const makeAs = (resourceType: string, name: string) => {
    if (resourceType === "stream") {
      return `/${toURLName(project.organization.name)}/${toURLName(project.name)}/${toURLName(name)}`;
    } else { // resourceType === "service"
      return `/${toURLName(project.organization.name)}/${toURLName(project.name)}/-/services/${toURLName(name)}`;
    }
  }

  return (
    <ContentContainer paper callToAction={cta}>
      <Table textSize="medium">
        <TableHead>
          <TableRow>
            <TableCell></TableCell>
            <TableCell>Name</TableCell>
            <Hidden smDown>
              <TableCell>Description</TableCell>
            </Hidden>
            <Hidden xsDown>
              <TableCell></TableCell>
            </Hidden>
          </TableRow>
        </TableHead>
        <TableBody>
          {Array.from(resources)
            .sort((a, b) => a.name.localeCompare(b.name))
            .map(({ type, name, description, resourceID }, idx) => (
              <TableLinkRow key={resourceID} href={makeHref(type, name)} as={makeAs(type, name)}>
                <TableCell className={classes.resourceTypeCell}>
                  <Paper
                    className={clsx(
                      classes.resourceTypePaper,
                      type === "stream" && classes.streamPaper,
                      type === "service" && classes.servicePaper
                    )}
                  >
                    <Typography variant="body2" className={classes.resourceTypePaperText}>
                      {type.charAt(0).toUpperCase() + type.slice(1)}
                    </Typography>
                  </Paper>
                </TableCell>
                <TableCell className={classes.nameCell}>{toURLName(name)}</TableCell>
                <Hidden smDown>
                  <TableCell className={classes.descriptionCell}>{description}</TableCell>
                </Hidden>
                <Hidden xsDown>
                  <ChipsCell resourceType={type} resourceID={resourceID} />
                </Hidden>
              </TableLinkRow>
            ))}
        </TableBody>
      </Table>
    </ContentContainer>
  );
};

export default ViewOverview;

// separate component to enable nested useTotalUsage
const ChipsCell: FC<{ resourceType: string; resourceID: string; }> = ({ resourceType, resourceID }) => {
  const classes = useStyles();
  if (resourceType === "stream") {
    const { data } = useTotalUsage(EntityKind.Stream, resourceID);
    if (!data) {
      return (
        <TableCell></TableCell>
      );
    }

    return (
      <TableCell align="right" className={classes.chipsCell}>
        <Grid container spacing={2} justify="flex-end">
          <Grid item>
            <Chip label={numbro(data.writeRecords).format(intFormat) + " records"} className={classes.chip} />
          </Grid>
          <Grid item>
            <Chip label={numbro(data.writeBytes).format(bytesFormat)} className={classes.chip} />
          </Grid>
        </Grid>
      </TableCell>
    );
  } else { // No chips for services for now
    return (
      <TableCell></TableCell>
    );
  }
}