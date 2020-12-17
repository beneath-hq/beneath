import numbro from "numbro";
import React, { FC } from "react";
import Moment from "react-moment";

import { EntityKind } from "apollo/types/globalTypes";
import { ProjectByOrganizationAndName_projectByOrganizationAndName } from "apollo/types/ProjectByOrganizationAndName";
import ContentContainer, { CallToAction } from "components/ContentContainer";
import { useTotalUsage } from "components/usage/util";
import { Table, TableBody, TableCell, TableHead, TableLinkRow, TableRow } from "components/Tables";
import { toURLName } from "lib/names";

const intFormat = { thousandSeparated: true };
const bytesFormat: numbro.Format = { base: "decimal", mantissa: 1, output: "byte" };

export interface ViewStreamsProps {
  project: ProjectByOrganizationAndName_projectByOrganizationAndName;
}

const ViewStreams: FC<ViewStreamsProps> = ({ project }) => {
  let cta: CallToAction | undefined;
  if (!project.streams?.length) {
    cta = {
      message: (
        <>
          We didn't find any streams in{" "}
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

  return (
    <ContentContainer paper callToAction={cta} maxWidth="md">
      <Table textSize="medium">
        <TableHead>
          <TableRow>
            <TableCell>Name</TableCell>
            <TableCell>Total records</TableCell>
            <TableCell>Total bytes</TableCell>
            {/* <TableCell align="right">Instances</TableCell> */}
            <TableCell align="right">Created</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {Array.from(project.streams)
            .sort((a, b) => a.name.localeCompare(b.name))
            .map(({ streamID, primaryStreamInstanceID, name, createdOn }, idx) => (
              <TableLinkRow
                key={streamID}
                href={
                  `/stream?organization_name=${toURLName(project.organization.name)}` +
                  `&project_name=${toURLName(project.name)}&stream_name=${toURLName(name)}`
                }
                as={`/${toURLName(project.organization.name)}/${toURLName(project.name)}/${toURLName(name)}`}
              >
                <TableCell>{toURLName(name)}</TableCell>
                <RecordsAndBytesCells skip={idx >= 25} streamInstanceID={primaryStreamInstanceID} />
                {/* <TableCell align="right">{instancesCreatedCount - instancesDeletedCount}</TableCell> */}
                <TableCell align="right">
                  <Moment fromNow>{createdOn}</Moment>
                </TableCell>
              </TableLinkRow>
            ))}
        </TableBody>
      </Table>
    </ContentContainer>
  );
};

export default ViewStreams;

// separate component to enable nested useTotalUsage
const RecordsAndBytesCells: FC<{ streamInstanceID: string | null; skip?: boolean }> = ({ streamInstanceID, skip }) => {
  if (skip || !streamInstanceID) {
    return (
      <>
        <TableCell>–</TableCell>
        <TableCell>–</TableCell>
      </>
    );
  }

  const { data } = useTotalUsage(EntityKind.StreamInstance, streamInstanceID);
  return (
    <>
      <TableCell>{data && numbro(data.writeRecords).format(intFormat)}</TableCell>
      <TableCell>{data && numbro(data.writeBytes).format(bytesFormat)}</TableCell>
    </>
  );
};
