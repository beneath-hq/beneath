import { useQuery } from "@apollo/react-hooks";
import React, { FC } from "react";

import { List, ListItem, ListItemAvatar, ListItemText, makeStyles, Typography } from "@material-ui/core";

import { QUERY_PROJECT_MEMBERS } from "../../apollo/queries/project";
import { ProjectByOrganizationAndName_projectByOrganizationAndName } from "../../apollo/types/ProjectByOrganizationAndName";
import { ProjectMembers, ProjectMembers_projectMembers, ProjectMembersVariables } from "../../apollo/types/ProjectMembers";
import { toURLName } from "../../lib/names";
import Avatar from "../Avatar";
import Loading from "../Loading";
import NextMuiLinkList from "../NextMuiLinkList";

interface ViewMembersProps {
  project: ProjectByOrganizationAndName_projectByOrganizationAndName;
}

const ViewMembers: FC<ViewMembersProps> = ({ project }) => {
  const { loading, error, data } = useQuery<ProjectMembers, ProjectMembersVariables>(QUERY_PROJECT_MEMBERS, {
    variables: { projectID: project.projectID },
  });

  if (loading) {
    return <Loading justify="center" />;
  }

  if (error || !data) {
    return <p>Error: {JSON.stringify(error)}</p>;
  }

  return (
    <>
      <List>
        {data.projectMembers.map((member) => (
          <ListItem
            key={member.userID}
            component={NextMuiLinkList}
            href={`/organization?organization_name=${toURLName(member.name)}`}
            as={`/${toURLName(member.name)}`}
            button
            disableGutters
          >
            {member.photoURL && (
              <ListItemAvatar>
                <Avatar size="list" label={member.displayName} src={member.photoURL} />
              </ListItemAvatar>
            )}
            <ListItemText
              primary={member.displayName || toURLName(member.name)}
              secondary={
                project.permissions.create || project.permissions.admin ?
                makePermissionsDescription(member) : undefined
              }
            />
          </ListItem>
        ))}
      </List>
    </>
  );
};

const makePermissionsDescription = (member: ProjectMembers_projectMembers) => {
  const res = [];
  if (member.view) {
    res.push("view streams");
  }
  if (member.create) {
    res.push("create and modify streams");
  }
  if (member.admin) {
    res.push("administrate project");
  }
  return `Permissions: ${res.join(", ")}`;
};

export default ViewMembers;
