import { useMutation, useQuery } from "@apollo/client";
import React, { FC, useState } from "react";

import ContentContainer, { CallToAction } from "components/ContentContainer";
import { UITable, UITableBody, UITableCell, UITableHead, UITableLinkCell, UITableRow } from "components/UITables";
import { toURLName } from "lib/names";
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
  IconButton,
  Typography,
} from "@material-ui/core";
import DeleteIcon from "@material-ui/icons/Delete";
import {
  TablePermissionsForService,
  TablePermissionsForServiceVariables,
} from "apollo/types/TablePermissionsForService";
import { QUERY_TABLE_PERMISSIONS_FOR_SERVICE } from "apollo/queries/service";
import {
  UpdateServiceTablePermissions,
  UpdateServiceTablePermissionsVariables,
} from "apollo/types/UpdateServiceTablePermissions";
import { UPDATE_SERVICE_TABLE_PERMISSIONS } from "apollo/queries/service";
import AddPermission from "./AddPermission";

export interface Props {
  serviceID: string;
  editable: boolean;
}

const ListPermissions: FC<Props> = ({ serviceID, editable }) => {
  const [showAddPermissionDialog, setShowAddPermissionDialog] = useState(false);
  const [showRevokePermissionDialog, setShowRevokePermissionDialog] = useState<string | undefined>(undefined);
  const { loading, error, data } = useQuery<TablePermissionsForService, TablePermissionsForServiceVariables>(
    QUERY_TABLE_PERMISSIONS_FOR_SERVICE,
    {
      variables: { serviceID },
    }
  );

  const [updateServiceTablePermissions, { loading: mutLoading }] = useMutation<
    UpdateServiceTablePermissions,
    UpdateServiceTablePermissionsVariables
  >(UPDATE_SERVICE_TABLE_PERMISSIONS);

  let cta: CallToAction | undefined;
  if (!data?.tablePermissionsForService.length) {
    cta = {
      message: `This service currently has no permissions for any resources`,
    };
    if (editable) {
      cta.buttons = [{ label: "Add table permission", onClick: () => setShowAddPermissionDialog(true) }];
    }
  }

  const addPermissionDialog = (
    <Dialog
      open={showAddPermissionDialog}
      onBackdropClick={() => setShowAddPermissionDialog(false)}
      maxWidth="sm"
      fullWidth
    >
      <DialogContent>
        <AddPermission serviceID={serviceID} onCompleted={() => setShowAddPermissionDialog(false)} />
      </DialogContent>
    </Dialog>
  );

  const revokePermissionDialog = (
    <Dialog open={!!showRevokePermissionDialog} onBackdropClick={() => setShowRevokePermissionDialog(undefined)}>
      <DialogTitle>Are you sure you want to delete this permission?</DialogTitle>
      <DialogContent>
        <DialogContentText>
          A service in production that depends on this permission will no longer work.
        </DialogContentText>
      </DialogContent>
      <DialogActions>
        <Button color="primary" autoFocus onClick={() => setShowRevokePermissionDialog(undefined)}>
          No, go back
        </Button>
        <Button
          color="primary"
          autoFocus
          onClick={() => {
            if (showRevokePermissionDialog) {
              const tableID = showRevokePermissionDialog;
              updateServiceTablePermissions({
                variables: { serviceID, tableID, read: false, write: false },
                update: (cache, { data }) => {
                  if (data && data.updateServiceTablePermissions) {
                    const queryData = cache.readQuery({
                      query: QUERY_TABLE_PERMISSIONS_FOR_SERVICE,
                      variables: { serviceID },
                    }) as any;
                    const filtered = queryData.tablePermissionsForService.filter(
                      (perms: any) => perms.tableID !== tableID
                    );
                    cache.writeQuery({
                      query: QUERY_TABLE_PERMISSIONS_FOR_SERVICE,
                      variables: { serviceID },
                      data: { tablePermissionsForService: filtered },
                    });
                  }
                },
              });
              setShowRevokePermissionDialog(undefined);
            }
          }}
        >
          Yes, I'm sure
        </Button>
      </DialogActions>
    </Dialog>
  );

  return (
    <>
      <Typography variant="h2" gutterBottom>
        Permissions
      </Typography>
      <Typography variant="body2">
        Services should have minimally viable permissions. The service will only be able to access the resources listed
        below.
      </Typography>
      <ContentContainer
        paper
        margin="normal"
        loading={loading}
        error={error && JSON.stringify(error)}
        callToAction={cta}
      >
        <UITable>
          <UITableHead>
            <UITableRow>
              <UITableCell>Resource</UITableCell>
              <UITableCell>Path</UITableCell>
              <UITableCell align="center">Read</UITableCell>
              <UITableCell align="center">Write</UITableCell>
              {editable && <UITableCell>Delete</UITableCell>}
            </UITableRow>
          </UITableHead>
          <UITableBody>
            {/* TODO: add an alphabetical sort of the path, using .sort() before .map() */}
            {data?.tablePermissionsForService.map((perms) => (
              <React.Fragment key={perms.tableID}>
                {perms.table && (
                  <UITableRow>
                    <UITableCell>Table</UITableCell>
                    <UITableLinkCell
                      href={`/table?organization_name=${toURLName(
                        perms.table.project.organization.name
                      )}&project_name=${toURLName(perms.table.project.name)}&table_name=${toURLName(perms.table.name)}`}
                      as={`/${toURLName(perms.table.project.organization.name)}/${toURLName(
                        perms.table.project.name
                      )}/table:${toURLName(perms.table.name)}`}
                    >
                      {`/${toURLName(perms.table.project.organization.name)}/${toURLName(
                        perms.table.project.name
                      )}/table:${toURLName(perms.table.name)}`}
                    </UITableLinkCell>
                    <UITableCell align="center">{perms.read && "✓"}</UITableCell>
                    <UITableCell align="center">{perms.write && "✓"}</UITableCell>
                    {editable && (
                      <UITableCell padding="checkbox" align="right">
                        <IconButton disabled={mutLoading} onClick={() => setShowRevokePermissionDialog(perms.tableID)}>
                          <DeleteIcon />
                        </IconButton>
                      </UITableCell>
                    )}
                  </UITableRow>
                )}
              </React.Fragment>
            ))}
          </UITableBody>
        </UITable>
        {addPermissionDialog}
        {revokePermissionDialog}
      </ContentContainer>
      {editable && !cta && (
        <Button variant="contained" onClick={() => setShowAddPermissionDialog(true)}>
          Add permission
        </Button>
      )}
    </>
  );
};

export default ListPermissions;
