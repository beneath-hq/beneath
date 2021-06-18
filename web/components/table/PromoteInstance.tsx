import { useMutation } from "@apollo/client";
import { Button, DialogActions, DialogContent, DialogContentText, DialogTitle } from "@material-ui/core";
import React, { FC } from "react";

import {
  QUERY_STREAM,
  QUERY_STREAM_INSTANCE,
  QUERY_STREAM_INSTANCES,
  UPDATE_STREAM_INSTANCE,
} from "apollo/queries/table";
import { TableByOrganizationProjectAndName_tableByOrganizationProjectAndName } from "apollo/types/TableByOrganizationProjectAndName";
import { UpdateTableInstance, UpdateTableInstanceVariables } from "apollo/types/UpdateTableInstance";
import { TableInstance } from "components/table/types";

export interface PromoteInstanceProps {
  table: TableByOrganizationProjectAndName_tableByOrganizationProjectAndName;
  instance: TableInstance;
  setOpenDialogID: (dialogID: "create" | "promote" | "delete" | null) => void;
}

const PromoteInstance: FC<PromoteInstanceProps> = ({ table, instance, setOpenDialogID }) => {
  const [updateTableInstance] = useMutation<UpdateTableInstance, UpdateTableInstanceVariables>(UPDATE_STREAM_INSTANCE, {
    onCompleted: () => setOpenDialogID(null),
  });

  return (
    <>
      <DialogTitle>Are you sure you want to make this instance the primary instance?</DialogTitle>
      <DialogContent>
        <DialogContentText>
          All prior versions and their data will be deleted. You won't be able to recover the data.
        </DialogContentText>
      </DialogContent>
      <DialogActions>
        <Button color="primary" autoFocus onClick={() => setOpenDialogID(null)}>
          No, go back
        </Button>
        <Button
          color="primary"
          autoFocus
          onClick={() =>
            updateTableInstance({
              variables: { input: { tableInstanceID: instance.tableInstanceID, makePrimary: true } },
              refetchQueries: [
                {
                  query: QUERY_STREAM,
                  variables: {
                    organizationName: table.project.organization.name,
                    projectName: table.project.name,
                    tableName: table.name,
                  },
                },
                {
                  query: QUERY_STREAM_INSTANCE,
                  variables: {
                    organizationName: table.project.organization.name,
                    projectName: table.project.name,
                    tableName: table.name,
                    version: instance.version,
                  },
                },
                {
                  query: QUERY_STREAM_INSTANCES,
                  variables: {
                    organizationName: table.project.organization.name,
                    projectName: table.project.name,
                    tableName: table.name,
                  },
                },
              ],
            })
          }
        >
          Yes, I'm sure
        </Button>
      </DialogActions>
    </>
  );
};

export default PromoteInstance;
