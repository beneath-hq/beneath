import avro from "avsc";
import React, { Component } from "react";
import Router from "next/router";
import { Mutation } from "react-apollo";

import Button from "@material-ui/core/Button";
import FormGroup from "@material-ui/core/FormGroup";
import TextField from "@material-ui/core/TextField";
import Typography from "@material-ui/core/Typography";
import { makeStyles } from "@material-ui/core";

import ExploreSidebar from "../../components/ExploreSidebar";
import SelectField from "../../components/SelectField";
import CheckboxField from "../../components/CheckboxField";
import Page from "../../components/Page";
import withMe from "../../hocs/withMe";
import { CREATE_EXTERNAL_STREAM } from "../../queries/stream";

const validateAvroSchema = (value) => {  
  if (typeof value !== "string") {
    return false;
  }

  try {
    const schema = JSON.parse(value);
    avro.Type.forSchema(schema, { noAnonymousTypes: true });
    return true;
  } catch {}

  return false;
};

const handleTabInput = (e) => {
  if (e.keyCode === 9) { // 9 = tab key
    let start = event.target.selectionStart;
    let end = event.target.selectionEnd;

    let value = event.target.value;
    value = value.substring(0, start) + "    " + value.substring(end);
    event.target.value = value;

    event.target.selectionStart = event.target.selectionEnd = start + 4;

    e.preventDefault();
    return false;
  }
};

const useStyles = makeStyles((theme) => ({
  title: {
    marginTop: theme.spacing(6),
  },
  submitButton: {
    marginTop: theme.spacing(3),
  },
}));

const NewStreamPage = ({ me }) => {
  const [values, setValues] = React.useState({
    projectId: "",
    name: "",
    description: "",
    avroSchema: "",
    batch: false,
    manual: true,
  });

  const handleChange = (name) => (event) => {
    setValues({ ...values, [name]: event.target.value });
  };

  const handleCheckboxChange = (name) => (event) => {
    setValues({ ...values, [name]: event.target.checked });
  };

  const classes = useStyles();
  return (
    <Page title="New Stream" maxWidth="md" sidebar={<ExploreSidebar />}>
      <Mutation mutation={CREATE_EXTERNAL_STREAM}
        update={(cache, { data: { createExternalStream } }) => {
        }}
        onCompleted={({ createExternalStream }) => {
          const stream = createExternalStream;
          Router.push(
            `/stream?name=${stream.name}&project_name=${stream.project.name}`,
            `/projects/${stream.project.name}/streams/${stream.name}`
          );
        }}
      >
        {(newStream, { loading, error }) => {
          const onSubmit = (e) => {
            e.preventDefault();
            newStream({ variables: values });
          };

          const isNameError = !!(error && error.message.match(/STREAM_PROJECT_NAME_UNIQUE/));

          return (
            <form onSubmit={onSubmit}>
              <Typography component="h2" variant="h2" gutterBottom className={classes.title}>Create stream</Typography>
              <SelectField id="project" label="Project" value={values.projectId} required
                helperText="Cannot be changed after creation" 
                onChange={handleChange("projectId")}
                options={me.user.projects.map((project) => ({
                  value: project.projectId,
                  label: project.name,
                }))}
              />
              <TextField id="name" label="Name" value={values.name}
                margin="normal" fullWidth required
                onChange={handleChange("name")}
                error={isNameError} helperText={isNameError && "Stream name already exists in project"}
              />
              <TextField id="description" label="Description" value={values.description}
                margin="normal" fullWidth required
                onChange={handleChange("description")}
              />
              <TextField id="schema" label="Avro Schema" value={values.avroSchema}
                margin="normal" multiline rows={10} rowsMax={1000} fullWidth required
                onChange={handleChange("avroSchema")}
                onKeyDown={handleTabInput}
              />
              <FormGroup row>
                <CheckboxField label="Batch" checked={values.batch} onChange={handleCheckboxChange("batch")} />
                <CheckboxField label="Allow manual entry" checked={values.manual} onChange={handleCheckboxChange("manual")} />
              </FormGroup>
              <Button type="submit" variant="outlined" color="primary" className={classes.submitButton}
                disabled={
                  loading
                  || !(values.projectId && values.projectId.length > 0)
                  || !(values.name.match(/^[_a-z][_\-a-z0-9]*$/))
                  || !(values.name && values.name.length >= 3 && values.name.length <= 40)
                  || !(values.description && values.description.length <= 256)
                  || !validateAvroSchema(values.avroSchema)
                }>
                Create stream
              </Button>
              {error && !isNameError && (
                <Typography variant="body1" color="error">An error occurred</Typography>
              )}
            </form>
          );
        }}
      </Mutation>
    </Page>
  );
};

export default withMe(NewStreamPage);
