import { useQuery } from "@apollo/client";
import { Table, TableBody, TableCell, TableHead, TableRow } from "@material-ui/core";
import React, { FC } from "react";

import { QUERY_PROJECT_MEMBERS } from "apollo/queries/project";
import { ProjectByOrganizationAndName_projectByOrganizationAndName } from "apollo/types/ProjectByOrganizationAndName";
import { ProjectMembers, ProjectMembersVariables } from "apollo/types/ProjectMembers";
import { toURLName } from "lib/names";
import Avatar from "components/Avatar";
import { NakedLink } from "components/Link";
import ContentContainer from "components/ContentContainer";

export interface ViewMembersProps {
  project: ProjectByOrganizationAndName_projectByOrganizationAndName;
}

const ViewMembers: FC<ViewMembersProps> = ({ project }) => {
  const { loading, error, data } = useQuery<ProjectMembers, ProjectMembersVariables>(QUERY_PROJECT_MEMBERS, {
    variables: { projectID: project.projectID },
  });

  return (
    <ContentContainer paper loading={loading} error={error && JSON.stringify(error)} note="Use the Beneath CLI to add members">
      <Table>
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
            <TableRow
              key={member.userID}
              component={NakedLink}
              href={`/organization?organization_name=${toURLName(member.name)}`}
              as={`/${toURLName(member.name)}`}
              hover
            >
              <TableCell>
                {member.photoURL && <Avatar size="list" label={member.displayName} src={member.photoURL} />}
              </TableCell>
              <TableCell>{toURLName(member.name)}</TableCell>
              <TableCell>{member.displayName}</TableCell>
              <TableCell align="center">{member.view && "✓"}</TableCell>
              <TableCell align="center">{member.create && "✓"}</TableCell>
              <TableCell align="center">{member.admin && "✓"}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </ContentContainer>
  );
};

export default ViewMembers;
