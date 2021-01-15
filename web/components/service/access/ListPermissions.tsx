import { useMutation, useQuery } from "@apollo/client";
import React, { FC, useState } from "react";

import ContentContainer, { CallToAction } from "components/ContentContainer";
import { Table, TableBody, TableCell, TableHead, TableLinkCell, TableLinkRow, TableRow } from "components/Tables";
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
  StreamPermissionsForService,
  StreamPermissionsForServiceVariables,
} from "apollo/types/StreamPermissionsForService";
import { QUERY_STREAM_PERMISSIONS_FOR_SERVICE } from "apollo/queries/service";
import {
  UpdateServiceStreamPermissions,
  UpdateServiceStreamPermissionsVariables,
} from "apollo/types/UpdateServiceStreamPermissions";
import { UPDATE_SERVICE_STREAM_PERMISSIONS } from "apollo/queries/service";
import AddPermission from "./AddPermission";

export interface Props {
  serviceID: string;
}

const ListPermissions: FC<Props> = ({ serviceID }) => {
  const [showAddPermissionDialog, setShowAddPermissionDialog] = useState(false);
  const [showRevokePermissionDialog, setShowRevokePermissionDialog] = useState<string | undefined>(undefined);
  const { loading, error, data } = useQuery<StreamPermissionsForService, StreamPermissionsForServiceVariables>(
    QUERY_STREAM_PERMISSIONS_FOR_SERVICE,
    {
      variables: { serviceID },
    }
  );

  const [updateServiceStreamPermissions, { loading: mutLoading }] = useMutation<
    UpdateServiceStreamPermissions,
    UpdateServiceStreamPermissionsVariables
  >(UPDATE_SERVICE_STREAM_PERMISSIONS);

  let cta: CallToAction | undefined;
  if (!data?.streamPermissionsForService.length) {
    cta = {
      message: `This service currently has no permissions for any resources`,
      buttons: [{ label: "Add stream permission", onClick: () => setShowAddPermissionDialog(true) }],
    };
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
              const streamID = showRevokePermissionDialog;
              updateServiceStreamPermissions({
                variables: { serviceID, streamID, read: false, write: false },
                update: (cache, { data }) => {
                  if (data && data.updateServiceStreamPermissions) {
                    const queryData = cache.readQuery({
                      query: QUERY_STREAM_PERMISSIONS_FOR_SERVICE,
                      variables: { serviceID },
                    }) as any;
                    const filtered = queryData.streamPermissionsForService.filter(
                      (perms: any) => perms.streamID !== streamID
                    );
                    cache.writeQuery({
                      query: QUERY_STREAM_PERMISSIONS_FOR_SERVICE,
                      variables: { serviceID },
                      data: { streamPermissionsForService: filtered },
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
        Services should have minimally viable permissions. This service will only be able to access the resources in
        this table.
      </Typography>
      <ContentContainer
        paper
        margin="normal"
        loading={loading}
        error={error && JSON.stringify(error)}
        callToAction={cta}
      >
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Resource</TableCell>
              <TableCell>Path</TableCell>
              <TableCell align="center">Read</TableCell>
              <TableCell align="center">Write</TableCell>
              <TableCell>Delete</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {/* TODO: add an alphabetical sort of the path, using .sort() before .map() */}
            {data?.streamPermissionsForService.map((perms) => (
              <React.Fragment key={perms.streamID}>
                {perms.stream && (
                  <TableRow>
                    <TableCell>Stream</TableCell>
                    <TableLinkCell
                      href={`/stream?organization_name=${toURLName(
                        perms.stream.project.organization.name
                      )}&project_name=${toURLName(perms.stream.project.name)}&stream_name=${toURLName(
                        perms.stream.name
                      )}`}
                      as={`/${toURLName(perms.stream.project.organization.name)}/${toURLName(
                        perms.stream.project.name
                      )}/stream:${toURLName(perms.stream.name)}`}
                    >
                      {`/${toURLName(perms.stream.project.organization.name)}/${toURLName(
                        perms.stream.project.name
                      )}/stream:${toURLName(perms.stream.name)}`}
                    </TableLinkCell>
                    <TableCell align="center">{perms.read && "✓"}</TableCell>
                    <TableCell align="center">{perms.write && "✓"}</TableCell>
                    <TableCell padding="checkbox" align="right">
                      <IconButton disabled={mutLoading} onClick={() => setShowRevokePermissionDialog(perms.streamID)}>
                        <DeleteIcon />
                      </IconButton>
                    </TableCell>
                  </TableRow>
                )}
              </React.Fragment>
            ))}
          </TableBody>
        </Table>
        {addPermissionDialog}
        {revokePermissionDialog}
      </ContentContainer>
      {!cta && (
        <Button variant="contained" onClick={() => setShowAddPermissionDialog(true)}>
          Add permission
        </Button>
      )}
    </>
  );
};

export default ListPermissions;
