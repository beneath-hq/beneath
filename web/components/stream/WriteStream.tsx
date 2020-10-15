import _ from "lodash";
import React, { FC } from "react";
import { Button, makeStyles, Typography } from "@material-ui/core";
import { Client } from "beneath";

import { StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName } from "../../apollo/types/StreamByOrganizationProjectAndName";
import { Schema } from "./schema";
import { Formik, Form, Field } from "formik";
import FormikTextField from "components/formik/TextField";
import SubmitControl from "components/forms/SubmitControl";
import { validateValue } from "./FilterField";
import { useToken } from "hooks/useToken";
import { useQuery } from "@apollo/client";
import { ProjectMembers, ProjectMembersVariables } from "apollo/types/ProjectMembers";
import { QUERY_PROJECT_MEMBERS } from "apollo/queries/project";
import useMe from "hooks/useMe";

const useStyles = makeStyles((theme) => ({
  firstTitle: {
    marginTop: theme.spacing(0),
    marginBottom: theme.spacing(2),
  },
  title: {
    marginTop: theme.spacing(4),
    marginBottom: theme.spacing(2),
  },
  errorMsg: {
    marginTop: theme.spacing(3),
  },
  button: {
    marginTop: theme.spacing(3),
    marginBotton: theme.spacing(2),
    marginRight: theme.spacing(3),
  },
}));

interface WriteStreamProps {
  stream: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName;
  instanceID: string;
  setWriteDialog: (writeDialog: boolean) => void;
}

const WriteStream: FC<WriteStreamProps> = ({ stream: streamMetadata, instanceID, setWriteDialog }) => {
  const classes = useStyles();
  const schema = new Schema(streamMetadata.avroSchema, streamMetadata.streamIndexes);
  const token = useToken();
  const me = useMe();
  const { error, data } = useQuery<ProjectMembers, ProjectMembersVariables>(QUERY_PROJECT_MEMBERS, {
    variables: { projectID: streamMetadata.project.projectID },
  });

  if (error || !data) {
    return <p>Error: {JSON.stringify(error)}</p>;
  }

  const user = data.projectMembers.find((user) => (user.userID === me?.personalUserID));

  if (!user?.create) {
    return (
      <>
        <Typography variant="h1" gutterBottom>
          Sorry, you can't do that
        </Typography>
        <Typography color="error" gutterBottom className={classes.errorMsg}>
          You don't have permission to write to this stream. Make sure you are logged in.
        </Typography>
        <Button onClick={() => setWriteDialog(false)} className={classes.button} variant="contained" color="primary">
          Go back
        </Button>
      </>
    );
  }

  const client = new Client({ secret: token || undefined });
  const stream = client.findStream({instanceID});

  const initialValues: any = {
    data: {},
    instanceID,
  };

  schema.columns.map((col) => {
    initialValues.data[col.name]= "";
  });

  const sanitize = (schema: Schema, data: any) => {
    // when a field is optional or a key, and an empty string, remove the field from the record object
    // note: a required string field, which isn't a key, CAN be an empty string, so we don't remove it
    schema.columns.map((col) => {
      if ((col.isNullable || col.isKey) && data[col.name] === "") {
        delete data[col.name];
      }
    });

    return;
  };

  return (
    <Formik
      initialValues={initialValues}
      onSubmit={async (values, actions) => {
        // clear any previous error
        actions.setStatus(null);

        sanitize(schema, values.data);
        const { writeID, error } = await stream.write(values.data);
        if (error) {
          actions.setStatus(error.message);
        }
        if (writeID) {
          // successfully wrote record
          setWriteDialog(false);
        }
      }}
    >
      {({ isSubmitting, status }) => (
        <Form>
          <Typography component="h2" variant="h1" gutterBottom>
            Write a record
          </Typography>
          {schema.columns.map((col, idx) => (
            <Field
              key={idx}
              name={"data." + col.name}
              // placeholder={getPlaceholder(col.inputType)} // placeholders aren't this easy
              validate={(val: string) => {
                return validateValue(col.inputType, val);
              }}
              component={FormikTextField}
              label={`${col.displayName} (${col.typeName})`}
              required={!col.isNullable}
            />
          ))}
          <SubmitControl label="Write record" errorAlert={status} disabled={isSubmitting} />
        </Form>
      )}
    </Formik>
  );
};

export default WriteStream;
