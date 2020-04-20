import React, { FC } from "react";

import { Container, Link as MUILink, makeStyles, Paper, Theme, Typography } from "@material-ui/core";

import { StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName } from "../../apollo/types/StreamByOrganizationProjectAndName";
import { GATEWAY_URL } from "../../lib/connection";
import { toURLName } from "../../lib/names";

import useMe from "../../hooks/useMe";
import CodeBlock from "../CodeBlock";
import LinkTypography from "../LinkTypography";
import VSpace from "../VSpace";

const useStyles = makeStyles((theme: Theme) => ({
  flashPaper: {
    padding: theme.spacing(2),
    marginBottom: theme.spacing(4),
  },
}));

interface StreamAPIProps {
  stream: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName;
}

const StreamAPI: FC<StreamAPIProps> = ({ stream }) => {
  const me = useMe();
  const classes = useStyles();
  return (
    <Container maxWidth={"md"}>
      <Paper className={classes.flashPaper}>
        <Typography variant="h3" gutterBottom>
          Note
        </Typography>
        <Typography variant="body1">
          {me && (
            <>
              To create a secret for connecting to Beneath, just head over to your{" "}
              <LinkTypography
                href={`/organization?organization_name=${me.user.username}&tab=secrets`}
                as={`/${me.user.username}/secrets`}
              >
                profile page
              </LinkTypography>
            </>
          )}
          {!me && (
            <>
              You'll first have to <LinkTypography href="/auth">create a user</LinkTypography> to get a secret for
              connecting to Beneath (don't worry, it's free and we won't share your data with anyone)
            </>
          )}
        </Typography>
      </Paper>
      <Typography variant="h3" gutterBottom>
        Python
      </Typography>
      <Typography variant="body2" paragraph>
        We provide a Python library that makes it easy to get data into e.g. a Jupyter notebook. Just copy and paste
        this snippet:
      </Typography>
      <CodeBlock language={"python"}>{`import beneath
client = beneath.Client()
df = await client.easy_read("${toURLName(stream.project.organization.name)}/${toURLName(stream.project.name)}/${toURLName(stream.name)}")`}</CodeBlock>
      <Typography variant="body2" paragraph>
        To run this code, you must first install our Python library with <code>pip install beneath</code> and
        authenticate by running <code>beneath auth SECRET</code> on the command-line.
      </Typography>
      <VSpace units={8} />

      <Typography variant="h3" gutterBottom>
        JavaScript
      </Typography>
      <Typography variant="body2" paragraph>
        You can query this stream directly from your front-end. Just copy and paste this snippet to get started. It's
        very important that you only use read-only secrets in your front-end.
      </Typography>
      <CodeBlock language={"javascript"}>{`fetch("${GATEWAY_URL}/projects/${toURLName(
        stream.project.name
      )}/streams/${toURLName(stream.name)}", {
  "Authorization": "Bearer SECRET",
  "Content-Type": "application/json",
})
.then(res => res.json())
.then(data => {
  // TODO: Add your logic here
  console.log(data)
})`}</CodeBlock>
      <Typography variant="body2" paragraph>
        Replace SECRET with a read-only secret (see note at the top of this page for instructions).
      </Typography>
      <VSpace units={8} />

      <Typography variant="h3" gutterBottom>
        BigQuery
      </Typography>
      <Typography variant="body2" paragraph>
        You can query this stream however you want using its public BigQuery dataset. Here's an example of how to query
        it from the BigQuery <LinkTypography href="https://console.cloud.google.com/bigquery">console</LinkTypography>.
      </Typography>
      <CodeBlock language={"sql"}>{`select * from \`${bigQueryName(stream)}\``}</CodeBlock>
      <Typography variant="body2" paragraph>
        Data is made available in BigQuery nearly in real time.
      </Typography>
      <VSpace units={8} />

      <Typography variant="h3" gutterBottom>
        REST API
      </Typography>
      <Typography variant="body2" paragraph>
        You can lookup records in this stream from anywhere using the REST API. Here's an example of how to query it
        from the command line:
      </Typography>
      <CodeBlock language={"bash"}>
        {`curl -H "Authorization: SECRET" ${GATEWAY_URL}/projects/${toURLName(stream.project.name)}/streams/${toURLName(
          stream.name
        )}`}
      </CodeBlock>
      <Typography variant="body2" paragraph>
        Replace SECRET with a read-only secret, which you can obtain from the "Secrets" tab on your profile page.
      </Typography>
      <VSpace units={4} />
    </Container>
  );
};

export default StreamAPI;

const bigQueryName = (stream: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName) => {
  const projectName = stream.project.name.replace(/-/g, "_");
  const streamName = stream.name.replace(/-/g, "_");
  // const idSlug = stream.currentStreamInstanceID ? stream.currentStreamInstanceID.slice(0, 8) : null;
  const idSlug = null;
  return `beneath.${projectName}.${streamName}${idSlug ? "_" + idSlug : ""}`;
};
