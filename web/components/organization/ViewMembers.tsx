import { useQuery } from "@apollo/client";
import React, { FC } from "react";

import { List, ListItem, ListItemAvatar, ListItemText, makeStyles } from "@material-ui/core";

import { QUERY_ORGANIZATION_MEMBERS } from "../../apollo/queries/organization";
import { OrganizationByName_organizationByName } from "../../apollo/types/OrganizationByName";
import {
  OrganizationMembers,
  OrganizationMembers_organizationMembers,
  OrganizationMembersVariables,
} from "../../apollo/types/OrganizationMembers";
import { toURLName } from "../../lib/names";
import Avatar from "../Avatar";
import Loading from "../Loading";
import { NakedLink } from "../Link";

export interface ViewMembersProps {
  organization: OrganizationByName_organizationByName;
}

const ViewMembers: FC<ViewMembersProps> = ({ organization }) => {
  const { loading, error, data } = useQuery<OrganizationMembers, OrganizationMembersVariables>(
    QUERY_ORGANIZATION_MEMBERS,
    {
      variables: { organizationID: organization.organizationID },
    }
  );

  if (loading) {
    return <Loading justify="center" />;
  }

  if (error || !data) {
    return <p>Error: {JSON.stringify(error)}</p>;
  }

  return (
    <>
      <List>
        {data.organizationMembers.map((member) => (
          <ListItem
            key={member.userID}
            component={NakedLink}
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
              secondary={makePermissionsDescription(organization, member)}
            />
          </ListItem>
        ))}
      </List>
    </>
  );
};

const makePermissionsDescription = (
  organization: OrganizationByName_organizationByName,
  member: OrganizationMembers_organizationMembers,
) => {
  const res = [];
  if (member.view) {
    res.push("view resources");
  }
  if (member.create) {
    res.push("create and modify resources");
  }
  if (member.admin) {
    res.push("administrate organization");
  }
  let desc = `Permissions: ${res.join(", ")}`;
  if (member.billingOrganizationID === organization.organizationID) {
    desc += " (this organization handles billing for the user)";
  } else {
    desc += " (this organization does not handle billing for the user)";
  }
  return desc;
};

export default ViewMembers;
