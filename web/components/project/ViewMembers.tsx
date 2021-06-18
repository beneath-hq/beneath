import { useQuery } from "@apollo/client";
import React, { FC, useState } from "react";

import { QUERY_PROJECT_MEMBERS } from "apollo/queries/project";
import { ProjectByOrganizationAndName_projectByOrganizationAndName } from "apollo/types/ProjectByOrganizationAndName";
import { ProjectMembers, ProjectMembersVariables } from "apollo/types/ProjectMembers";
import Avatar from "components/Avatar";
import ContentContainer from "components/ContentContainer";
import { UITable, UITableBody, UITableCell, UITableHead, UITableLinkRow, UITableRow } from "components/UITables";
import { toURLName } from "lib/names";
import AddProjectMember from "./AddProjectMember";
import { Button, Dialog, DialogContent, Grid } from "@material-ui/core";

export interface ViewMembersProps {
  project: ProjectByOrganizationAndName_projectByOrganizationAndName;
}

const ViewMembers: FC<ViewMembersProps> = ({ project }) => {
  const [showAddMember, setShowAddMember] = useState(false);
  const { loading, error, data } = useQuery<ProjectMembers, ProjectMembersVariables>(QUERY_PROJECT_MEMBERS, {
    variables: { projectID: project.projectID },
  });

  return (
    <>
      <ContentContainer
        paper
        loading={loading}
        error={error && JSON.stringify(error)}
        note={
          !project.permissions.admin ? undefined : (
            <Grid container>
              <Grid item xs />
              <Grid item>
                <Button variant="contained" color="primary" onClick={() => setShowAddMember(true)}>
                  Add member
                </Button>
              </Grid>
            </Grid>
          )
        }
      >
        <UITable textSize="medium">
          <UITableHead>
            <UITableRow>
              <UITableCell padding="checkbox"></UITableCell>
              <UITableCell>Username</UITableCell>
              <UITableCell>Name</UITableCell>
              <UITableCell align="center">View</UITableCell>
              <UITableCell align="center">Create</UITableCell>
              <UITableCell align="center">Admin</UITableCell>
            </UITableRow>
          </UITableHead>
          <UITableBody>
            {data?.projectMembers.map((member) => (
              <UITableLinkRow
                key={member.userID}
                href={`/organization?organization_name=${toURLName(member.name)}`}
                as={`/${toURLName(member.name)}`}
              >
                <UITableCell>
                  {member.photoURL && <Avatar size="list" label={member.displayName} src={member.photoURL} />}
                </UITableCell>
                <UITableCell>{toURLName(member.name)}</UITableCell>
                <UITableCell>{member.displayName}</UITableCell>
                <UITableCell align="center">{member.view && "✓"}</UITableCell>
                <UITableCell align="center">{member.create && "✓"}</UITableCell>
                <UITableCell align="center">{member.admin && "✓"}</UITableCell>
              </UITableLinkRow>
            ))}
          </UITableBody>
        </UITable>
      </ContentContainer>
      <Dialog open={showAddMember} onBackdropClick={() => setShowAddMember(false)}>
        <DialogContent>
          <AddProjectMember projectID={project.projectID} onCompleted={() => setShowAddMember(false)} />
        </DialogContent>
      </Dialog>
    </>
  );
};

export default ViewMembers;
