import { useQuery } from "@apollo/client";
import React, { FC, useState } from "react";

import { QUERY_ORGANIZATION_MEMBERS } from "apollo/queries/organization";
import { OrganizationByName_organizationByName } from "apollo/types/OrganizationByName";
import { OrganizationMembers, OrganizationMembersVariables } from "apollo/types/OrganizationMembers";
import Avatar from "components/Avatar";
import ContentContainer from "components/ContentContainer";
import { NakedLink } from "components/Link";
import { UITable, UITableBody, UITableCell, UITableHead, UITableLinkRow, UITableRow } from "components/UITables";
import { toURLName } from "lib/names";
import { Button, Dialog, DialogContent, Grid } from "@material-ui/core";
import AddOrganizationMember from "./AddOrganizationMember";

export interface ViewMembersProps {
  organization: OrganizationByName_organizationByName;
}

const ViewMembers: FC<ViewMembersProps> = ({ organization }) => {
  const [showAddMember, setShowAddMember] = useState(false);
  const { loading, error, data } = useQuery<OrganizationMembers, OrganizationMembersVariables>(
    QUERY_ORGANIZATION_MEMBERS,
    {
      variables: { organizationID: organization.organizationID },
    }
  );

  return (
    <>
      <ContentContainer
        paper
        loading={loading}
        error={error && JSON.stringify(error)}
        note={
          !(organization.__typename === "PrivateOrganization" && organization.permissions.admin) ? undefined : (
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
              <UITableCell align="center">Billing handled by</UITableCell>
            </UITableRow>
          </UITableHead>
          <UITableBody>
            {data?.organizationMembers.map((member) => (
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
                <UITableCell align="center">
                  {member.billingOrganizationID === organization.organizationID
                    ? "This organization"
                    : "Other organization"}
                </UITableCell>
              </UITableLinkRow>
            ))}
          </UITableBody>
        </UITable>
      </ContentContainer>
      <Dialog open={showAddMember} onBackdropClick={() => setShowAddMember(false)}>
        <DialogContent>
          <AddOrganizationMember
            organizationID={organization.organizationID}
            onCompleted={() => setShowAddMember(false)}
          />
        </DialogContent>
      </Dialog>
    </>
  );
};

export default ViewMembers;
