import Link from "next/link";
import React, { FC } from "react";

import { makeStyles, Theme } from "@material-ui/core";
import Container from "@material-ui/core/Container";
import MUILink from "@material-ui/core/Link";
import Typography from "@material-ui/core/Typography";

import { QueryStream, QueryStream_stream } from "../../apollo/types/QueryStream";
import { GATEWAY_URL } from "../../lib/connection";
import { Schema } from "./schema";

import CodeBlock from "../CodeBlock";
import VSpace from "../VSpace";

const useStyles = makeStyles((theme: Theme) => ({
  link: {
    cursor: "pointer",
  },
}));

const StreamAPI: FC<QueryStream> = ({ stream }) => {
  const classes = useStyles();
  return (
    <Container maxWidth={"md"}>
      <Typography variant="h3" gutterBottom>
        Python
      </Typography>
      <Typography variant="body2" paragraph>
        We provide a Python library that makes it easy to get data into e.g. a Jupyter notebook. Just copy and paste
        this snippet:
      </Typography>
      <CodeBlock language={"python"}>{`import beneath
client = Client(TOKEN)
stream = client.Stream(project="${stream.project.name}", stream="${stream.name}")
df = stream.load_all().to_dataframe()`}</CodeBlock>
      <Typography variant="body2" paragraph>
        Replace TOKEN with a read-only key, which you can obtain{" "}
        <Link href={"/user?id=me&tab=keys"} as={"/users/me/keys"}>
          <MUILink className={classes.link}>here</MUILink>
        </Link>
        . You must first install our Python library with <code>pip install beneath</code>.
      </Typography>
      <VSpace units={4} />

      <Typography variant="h3" gutterBottom>
        JavaScript
      </Typography>
      <Typography variant="body2" paragraph>
        You can query this stream directly from your front-end. Just copy and paste this snippet to get started. It's
        very important that you only use read-only keys in your front-end.
      </Typography>
      <CodeBlock language={"javascript"}>{`fetch("${GATEWAY_URL}/projects/${stream.project.name}/streams/${
        stream.name
      }", {
  "Authorization": "Bearer TOKEN",
  "Content-Type": "application/json",
})
.then(res => res.json())
.then(data => {
  // TODO: Add your logic here
  console.log(data)
})`}</CodeBlock>
      <Typography variant="body2" paragraph>
        Replace TOKEN with a read-only key, which you can obtain{" "}
        <Link href={"/user?id=me&tab=keys"} as={"/users/me/keys"}>
          <MUILink className={classes.link}>here</MUILink>
        </Link>
        .
      </Typography>
      <VSpace units={4} />

      <Typography variant="h3" gutterBottom>
        BigQuery
      </Typography>
      <Typography variant="body2" paragraph>
        You can query this stream however you want using its public BigQuery dataset. Here's an example of how to
        query it from the BigQuery{" "}
        <MUILink href="https://console.cloud.google.com/bigquery" className={classes.link}>
          console
        </MUILink>
        .
      </Typography>
      <CodeBlock language={"sql"}>{`select * from \`${bigQueryName(stream)}\``}</CodeBlock>
      <Typography variant="body2" paragraph>
        Data is made available in BigQuery nearly in real time.
      </Typography>
      <VSpace units={4} />

      <VSpace units={4} />
      <Typography variant="h3" gutterBottom>
        REST API
      </Typography>
      <Typography variant="body2" paragraph>
        You can lookup records in this stream from anywhere using the REST API. Here's an example of how to query it
        from the command line:
      </Typography>
      <CodeBlock language={"bash"}>
        {`curl -H "Authorization: TOKEN" ${GATEWAY_URL}/projects/${stream.project.name}/streams/${stream.name}`}
      </CodeBlock>
      <Typography variant="body2" paragraph>
        Replace TOKEN with a read-only key, which you can obtain{" "}
        <Link href={"/user?id=me&tab=keys"} as={"/users/me/keys"}>
          <MUILink className={classes.link}>here</MUILink>
        </Link>
        .
      </Typography>
      <VSpace units={4} />
    </Container>
  );
};

export default StreamAPI;

const bigQueryName = (stream: QueryStream_stream) => {
  const projectName = stream.project.name.replace(/-/g, "_");
  const streamName = stream.name.replace(/-/g, "_");
  const idSlug = stream.currentStreamInstanceID ? stream.currentStreamInstanceID.slice(0, 8) : null;
  return `beneathcrypto.${projectName}.${streamName}${idSlug ? "_" + idSlug : ""}`;
};
