import { ApolloQueryResult } from "@apollo/client";
import { makeStyles, Paper, Typography, Theme, Box } from "@material-ui/core";
import _ from "lodash";
import React, { FC } from "react";

import { Link } from "components/Link";
import { StreamByOrganizationProjectAndName } from "apollo/types/StreamByOrganizationProjectAndName";
import { toURLName } from "lib/names";
import SchemaView from "./SchemaView";

const useStyles = makeStyles((_: Theme) => ({
  message: {
    margin: "2rem 0",
  },
  description: {
    marginTop: "0.3rem",
  },
  streamPreviewPaper: {
    padding: "12px",
  },
}));

export interface StreamPreviewProps {
  result: ApolloQueryResult<StreamByOrganizationProjectAndName>;
}

export const StreamPreview: FC<StreamPreviewProps> = ({ result: { data, error, errors } }) => {
  let errorMsg = null;
  if (error || errors || !data) {
    errorMsg = error ? error.message : errors && errors.length > 0 ? errors[0].message : "Unknown error";
  }

  const stream = data?.streamByOrganizationProjectAndName;
  const orgName = toURLName(stream?.project.organization.name || "");
  const projName = toURLName(stream?.project.name || "");

  const classes = useStyles();
  return (
    <Paper className={classes.streamPreviewPaper}>
      {errorMsg && (
        <Typography className={classes.message} variant="h4" color="error" align="center">
          {errorMsg}
        </Typography>
      )}
      {stream && (
        <>
          <Typography variant="subtitle2">Stream schema</Typography>
          <Link
            color="textPrimary"
            underline="none"
            variant="h3"
            gutterBottom
            href={`/stream?organization_name=${orgName}&project_name=${projName}&stream_name=${stream.name}`}
            as={`/${orgName}/${projName}/stream:${stream.name}`}
          >
            {orgName}/{projName}/{toURLName(stream.name)}
          </Link>
          <Typography color="textSecondary" variant="body2" gutterBottom className={classes.description}>
            {stream.description}
          </Typography>
          <Box mt={1.5}>
            <SchemaView stream={stream} />
          </Box>
        </>
      )}
    </Paper>
  );
};

export default StreamPreview;
