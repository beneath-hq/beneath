import { useQuery } from "@apollo/client";
import React, { FC } from "react";

import { QUERY_PROJECT_MEMBERS } from "apollo/queries/project";
import { ProjectByOrganizationAndName_projectByOrganizationAndName } from "apollo/types/ProjectByOrganizationAndName";
import { ProjectMembers, ProjectMembersVariables } from "apollo/types/ProjectMembers";
import Avatar from "components/Avatar";
import ContentContainer from "components/ContentContainer";
import { Table, TableBody, TableCell, TableHead, TableLinkRow, TableRow } from "components/Tables";
import { toURLName } from "lib/names";

export interface ViewMembersProps {
  project: ProjectByOrganizationAndName_projectByOrganizationAndName;
}

const ViewMembers: FC<ViewMembersProps> = ({ project }) => {
  const { loading, error, data } = useQuery<ProjectMembers, ProjectMembersVariables>(QUERY_PROJECT_MEMBERS, {
    variables: { projectID: project.projectID },
  });

  return (
    <ContentContainer
      paper
      loading={loading}
      error={error && JSON.stringify(error)}
      note="Use the Beneath CLI to add members"
    >
      <Table textSize="medium">
        <TableHead>
          <TableRow>
            <TableCell padding="checkbox"></TableCell>
            <TableCell>Username</TableCell>
            <TableCell>Name</TableCell>
            <TableCell align="center">View</TableCell>
            <TableCell align="center">Create</TableCell>
            <TableCell align="center">Admin</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          {data?.projectMembers.map((member) => (
            <TableLinkRow
              key={member.userID}
              href={`/organization?organization_name=${toURLName(member.name)}`}
              as={`/${toURLName(member.name)}`}
            >
              <TableCell>
                {member.photoURL && <Avatar size="list" label={member.displayName} src={member.photoURL} />}
              </TableCell>
              <TableCell>{toURLName(member.name)}</TableCell>
              <TableCell>{member.displayName}</TableCell>
              <TableCell align="center">{member.view && "✓"}</TableCell>
              <TableCell align="center">{member.create && "✓"}</TableCell>
              <TableCell align="center">{member.admin && "✓"}</TableCell>
            </TableLinkRow>
          ))}
        </TableBody>
      </Table>
    </ContentContainer>
  );
};

export default ViewMembers;
