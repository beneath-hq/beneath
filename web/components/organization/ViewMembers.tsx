import { useQuery } from "@apollo/client";
import React, { FC, useState } from "react";

import { QUERY_ORGANIZATION_MEMBERS } from "apollo/queries/organization";
import { OrganizationByName_organizationByName } from "apollo/types/OrganizationByName";
import { OrganizationMembers, OrganizationMembersVariables } from "apollo/types/OrganizationMembers";
import Avatar from "components/Avatar";
import ContentContainer from "components/ContentContainer";
import { NakedLink } from "components/Link";
import { Table, TableBody, TableCell, TableHead, TableLinkRow, TableRow } from "components/Tables";
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
        <Table textSize="medium">
          <TableHead>
            <TableRow>
              <TableCell padding="checkbox"></TableCell>
              <TableCell>Username</TableCell>
              <TableCell>Name</TableCell>
              <TableCell align="center">View</TableCell>
              <TableCell align="center">Create</TableCell>
              <TableCell align="center">Admin</TableCell>
              <TableCell align="center">Billing handled by</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {data?.organizationMembers.map((member) => (
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
                <TableCell align="center">
                  {member.billingOrganizationID === organization.organizationID
                    ? "This organization"
                    : "Other organization"}
                </TableCell>
              </TableLinkRow>
            ))}
          </TableBody>
        </Table>
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
