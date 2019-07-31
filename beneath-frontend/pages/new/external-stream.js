import React, { Component } from "react";
import Router from "next/router";
import { Mutation } from "react-apollo";

import Button from "@material-ui/core/Button";
import FormGroup from "@material-ui/core/FormGroup";
import Link from "@material-ui/core/Link";
import TextField from "@material-ui/core/TextField";
import Typography from "@material-ui/core/Typography";
import { makeStyles } from "@material-ui/core";

import ExploreSidebar from "../../components/ExploreSidebar";
import SelectField from "../../components/SelectField";
import CheckboxField from "../../components/CheckboxField";
import Page from "../../components/Page";
import withMe from "../../hocs/withMe";
import { QUERY_PROJECT } from "../../apollo/queries/project";
import { CREATE_EXTERNAL_STREAM } from "../../apollo/queries/stream";

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
  submitButton: {
    marginTop: theme.spacing(3),
  },
}));

const NewStreamPage = ({ me }) => {
  const [values, setValues] = React.useState({
    projectID: "",
    description: "",
    schema: "",
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
    <Page title="New External Stream" sidebar={<ExploreSidebar />} maxWidth="md" contentMarginTop="normal">
      <Mutation mutation={CREATE_EXTERNAL_STREAM}
        refetchQueries={({ data: { createExternalStream } }) => {
          const stream = createExternalStream;
          return [{ query: QUERY_PROJECT, variables: { name: stream.project.name } }];
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

          const isNameError = !!(error && error.message.match(/IDX_UQ_STREAMS_NAME_PROJECT/));

          return (
            <form onSubmit={onSubmit}>
              <Typography component="h2" variant="h2" gutterBottom>
                Create external stream
              </Typography>
              <SelectField
                id="project"
                label="Project"
                value={values.projectID}
                options={me.user.projects.map((project) => ({
                  value: project.projectID,
                  label: project.name,
                }))}
                onChange={handleChange("projectID")}
                required
                margin="normal"
                helperText="You can't change this later, so make sure you get it right"
              />
              <TextField
                id="description"
                label="Description"
                value={values.description}
                margin="normal"
                fullWidth
                required
                onChange={handleChange("description")}
                helperText="Help people understand what data this stream will contain"
              />
              <TextField
                id="schema"
                label="Schema"
                value={values.schema}
                margin="normal"
                multiline rows={10} rowsMax={1000}
                fullWidth
                required
                onChange={handleChange("schema")}
                onKeyDown={handleTabInput}
                helperText={<Typography variant="caption">Specify the format of data on the stream as an <Link target="_blank" href="https://docs.oracle.com/database/nosql-12.1.3.0/GettingStartedGuide/avroschemas.html">GraphQL schema</Link></Typography>}
              />
              <FormGroup>
                <CheckboxField 
                  label="Batch"
                  checked={values.batch}
                  onChange={handleCheckboxChange("batch")}
                  margin="normal"
                  helperText="If marked, data can only be written to the stream once. Use for reference data that never changes."
                />
                <CheckboxField 
                  label="Allow manual entry"
                  checked={values.manual}
                  onChange={handleCheckboxChange("manual")}
                  margin="normal"
                  helperText="Enable writing data to the stream through the UI" 
                />
              </FormGroup>
              <Button type="submit" variant="outlined" color="primary" className={classes.submitButton}
                disabled={
                  loading
                  || !(values.projectID && values.projectID.length > 0)
                  || !(values.description && values.description.length <= 255)
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
