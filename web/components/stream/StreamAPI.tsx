import React, { FC, useState } from "react";
import { Alert, TabContext } from "@material-ui/lab";
import { Container, Grid, Link as MUILink, makeStyles, Tab, Tabs, Theme, Typography } from "@material-ui/core";

import { StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName } from "../../apollo/types/StreamByOrganizationProjectAndName";
import { GATEWAY_URL } from "../../lib/connection";
import { toURLName } from "../../lib/names";
import useMe from "../../hooks/useMe";
import CodeBlock from "../CodeBlock";
import { Link } from "../Link";
import VSpace from "../VSpace";

const useStyles = makeStyles((theme: Theme) => ({
  container: {
    padding: "0px",
  },
  tab: {
    fontSize: "14px",
    padding: "14px",
  },
}));

interface StreamAPIProps {
  stream: StreamByOrganizationProjectAndName_streamByOrganizationProjectAndName;
}

const StreamAPI: FC<StreamAPIProps> = ({ stream }) => {
  const languageTabs = ["Python", "Javascript", "React", "SQL", "cURL"];
  const pythonTabs = ["Setup", "Reading", "Writing", "Pipelines"];
  const javascriptTabs = ["Setup", "Reading"];
  const reactTabs = ["Setup", "Reading"];
  
  const me = useMe();
  const classes = useStyles();
  const [language, setLanguage] = useState(languageTabs[0]);
  const [pythonDetail, setPythonDetail] = useState(pythonTabs[1]);
  const [javascriptDetail, setJavascriptDetail] = useState(javascriptTabs[1]);
  const [reactDetail, setReactDetail] = useState(reactTabs[1]);

  return (
    <Container maxWidth="md" className={classes.container}>
      <Alert severity="info">
        {me && (
          <>
            To create a secret for connecting to Beneath, head to the{" "}
            <Link
              href={`/organization?organization_name=${toURLName(me.name)}&tab=secrets`}
              as={`/${toURLName(me.name)}/-/secrets`}
            >
              secrets page
            </Link>
          </>
        )}
        {!me && (
          <>
            You'll first have to <Link href="/-/auth">login or signup</Link> to get a secret for
            connecting to Beneath (don't worry, it's free and we won't share your data with anyone)
          </>
        )}
      </Alert>
      <VSpace units={2} />
      <Grid container justify="space-between">
        <Grid item xs={12} md={6}>
          <TabContext value={language}>
            <Tabs
              value={language}
              onChange={(_, value) => setLanguage(value)}
              variant="scrollable"
              scrollButtons="auto"
            >
              {languageTabs.map((tab) => (
                <Tab
                  key={tab}
                  label={tab}
                  value={tab}
                  className={classes.tab}
                />
              ))}
            </Tabs>
          </TabContext>
        </Grid>
        {language === "Python" && (
          <Grid item>
            <TabContext value={pythonDetail}>
              <Tabs
                value={pythonDetail}
                onChange={(_, value) => setPythonDetail(value)}
                variant="scrollable"
                scrollButtons="auto"
              >
                {pythonTabs.map((tab) => (
                  <Tab
                    key={tab}
                    label={tab}
                    value={tab}
                    className={classes.tab}
                  />
                ))}
              </Tabs>
            </TabContext>
          </Grid>
        )}
        {language === "Javascript" && (
          <Grid item>
            <TabContext value={javascriptDetail}>
              <Tabs
                value={javascriptDetail}
                onChange={(_, value) => setJavascriptDetail(value)}
                variant="scrollable"
                scrollButtons="auto"
              >
                {javascriptTabs.map((tab) => (
                  <Tab
                    key={tab}
                    label={tab}
                    value={tab}
                    className={classes.tab}
                  />
                ))}
              </Tabs>
            </TabContext>
          </Grid>
        )}
        {language === "React" && (
          <Grid item>
            <TabContext value={reactDetail}>
              <Tabs
                value={reactDetail}
                onChange={(_, value) => setReactDetail(value)}
                variant="scrollable"
                scrollButtons="auto"
              >
                {reactTabs.map((tab) => (
                  <Tab
                    key={tab}
                    label={tab}
                    value={tab}
                    className={classes.tab}
                  />
                ))}
              </Tabs>
            </TabContext>
          </Grid>
        )}
      </Grid>
      <VSpace units={4} />
        <Grid item>
          {language === "Python" && pythonDetail === "Setup" && (
            <>
              <Typography variant="body1" paragraph>
                Install the Beneath Python library:
              </Typography>
              <CodeBlock language={"bash"}>
{`pip install beneath`}
              </CodeBlock>
              <VSpace units={2} />
              <Typography variant="body1" paragraph>
                Authenticate your environment:
              </Typography>
              <CodeBlock language={"bash"}>
{`beneath auth SECRET`}
              </CodeBlock>
            </>
          )}
          {language === "Python" && pythonDetail === "Reading" && (
            <>
              <Typography variant="body1" paragraph>
                From a Python script or notebook:
              </Typography>
              <CodeBlock language={"python"}>
{`import beneath

df = await beneath.easy_read("${toURLName(stream.project.organization.name)}/${toURLName(stream.project.name)}/${toURLName(stream.name)}")`}
              </CodeBlock>
            </>
          )}
          {language === "Python" && pythonDetail === "Writing" && (
            <>
              <Typography variant="body1" paragraph>
                From a Python script or notebook:
              </Typography>
              <CodeBlock language={"python"}>
{`import beneath
client = beneath.Client()

schema = """
${stream.schema}
"""

stream = await client.create_stream(stream_path="${toURLName(stream.project.organization.name)}/${toURLName(stream.project.name)}/${toURLName(stream.name)}", schema=schema)
instance = await stream.create_instance(version=VERSION)
async with instance.writer() as w:
    await w.write(LIST_OF_RECORDS)`}
              </CodeBlock>
            </>
          )}
          {language === "Python" && pythonDetail === "Pipelines" && (
            <>
              <Typography paragraph>
                Beneath Pipelines make it easy to do stream processing on any Beneath data stream.
              </Typography>
              <Typography paragraph>
                First, create a new Python file for your pipeline logic. Choose one from the "Consume," "Derive," or "Advanced" options below.
              </Typography>
              <Typography variant="subtitle1" gutterBottom>
                Consume: apply a user-defined function
              </Typography>
              <CodeBlock language={"python"} title={"your_pipeline.py"}>
{`import beneath

async def consume_fn(record):
  # YOUR LOGIC HERE

beneath.easy_consume_stream(
  input_stream_path="${toURLName(stream.project.organization.name)}/${toURLName(stream.project.name)}/${toURLName(stream.name)}",
  consume_fn=consume_fn,
)`}
              </CodeBlock>
              <VSpace units={2} />
              <Typography variant="subtitle1" gutterBottom>
                Derive: apply a user-defined function and write results to a new stream
              </Typography>
              <CodeBlock language={"python"} title={"your_pipeline.py"}>
{`import beneath

OUTPUT_SCHEMA="""
YOUR OUTPUT SCHEMA HERE
"""

async def apply_fn(record):
  # YOUR LOGIC HERE
  yield new_record

beneath.easy_derive_stream(
  input_stream_path="${toURLName(stream.project.organization.name)}/${toURLName(stream.project.name)}/${toURLName(stream.name)}",
  apply_fn=apply_fn,
  output_stream_path="USER/PROJECT/YOUR_NEW_STREAM_NAME",
  output_stream_schema=OUTPUT_SCHEMA
)`}
              </CodeBlock>
              <VSpace units={2} />
              <Typography variant="subtitle1" gutterBottom>
                Advanced: check out the full docs to see what you can do with an advanced pipeline
              </Typography>
              <VSpace units={2} />
              <Typography variant="body1" paragraph>
                Second, <Link href={"/-/create/service"}>create a service</Link>, then stage your pipeline:
              </Typography>
              <CodeBlock language={"bash"}>
{`python your_pipeline.py stage USERNAME/PROJECT/SERVICE`}
              </CodeBlock>
              <VSpace units={2} />
              <Typography variant="body1" paragraph>
                Lastly, run your pipeline:
              </Typography>
              <CodeBlock language={"bash"}>
{`python your_pipeline.py run USERNAME/PROJECT/SERVICE`}
              </CodeBlock>
            </>
          )}
          {language === "Javascript" && javascriptDetail === "Setup" && (
            <>
              <Typography variant="body1" paragraph>
                Install the Javascript library with npm:
              </Typography>
              <CodeBlock language={"bash"}>
{`npm install beneath`}
              </CodeBlock>
              <VSpace units={2} />
              <Typography variant="body1" paragraph>
                Or install with yarn:
              </Typography>
              <CodeBlock language={"bash"}>
{`yarn install beneath`}
              </CodeBlock>
            </>
          )}
          {language === "Javascript" && javascriptDetail === "Reading" && (
            <>
              <Typography variant="body1" paragraph>
                You can query this stream directly from your frontend. It's very important to use permissioned "read-only" secrets in your frontend code.
              </Typography>
              <CodeBlock language={"javascript"}>
                {`fetch("${GATEWAY_URL}/v1/${toURLName(
                stream.project.organization.name
              )}/${toURLName(stream.project.name)}/${toURLName(stream.name)}", {
    "Authorization": "Bearer SECRET",
    "Content-Type": "application/json",
  })
  .then(res => res.json())
  .then(data => {
    // TODO: Add your logic here
    console.log(data)
  })`}
              </CodeBlock>
            </>
          )}
          {language === "React" && reactDetail === "Setup" && (
            <>
              <Typography variant="body1" paragraph>
                Install the React library with npm:
              </Typography>
              <CodeBlock language={"bash"}>
{`npm install beneath-react`}
              </CodeBlock>
              <VSpace units={2} />
              <Typography variant="body1" paragraph>
                Or install with yarn:
              </Typography>
              <CodeBlock language={"bash"}>
{`yarn install beneath-react`}
              </CodeBlock>
            </>
          )}
          {language === "React" && reactDetail === "Reading" && (
            <>
              <Typography variant="body1" paragraph>
                You can query this stream directly from your frontend. It's very important to use permissioned "read-only" secrets in your frontend code.
              </Typography>
              <CodeBlock language={"javascript"}>
{`import { useRecords } from "beneath-react";

const App = () => {
  const { records, error, loading } = useRecords({
    stream: "${toURLName(stream.project.organization.name)}/${toURLName(stream.project.name)}/${toURLName(stream.name)}",
    query: {type: "log", peek: "true"},
    pageSize: 500,
  });
}`}
              </CodeBlock>
            </>
          )}
          {language === "SQL" && (
            <>
              <Typography variant="body1" paragraph>
                You can query this stream in the <Link href="/-/sql">Beneath SQL Editor</Link> like this:
              </Typography>
              <CodeBlock language={"sql"}>{`select * from \`${toURLName(stream.project.organization.name)}/${toURLName(stream.project.name)}/${toURLName(stream.name)}\``}</CodeBlock>
              <VSpace units={2} />
              <Typography variant="body1">
                Additionally, you can connect any business intelligence tool that you'd like, such as Metabase, Tableau, Mode, etc. Please reach out if you'd to do this, as it currently takes just a little hand-holding.
              </Typography>
            </>
          )}
          {language === "cURL" && (
            <>
              <Typography variant="body1" paragraph>
                You can query this data stream from anywhere using the REST API. Here's an example using cURL from the command line:
              </Typography>
              <CodeBlock language={"bash"}>
                {`curl -H "Authorization: Bearer SECRET" ${GATEWAY_URL}/v1/${toURLName(stream.project.organization.name)}/${toURLName(
                  stream.project.name
                )}/${toURLName(stream.name)}`}
              </CodeBlock>
            </>
          )}
        </Grid>
    </Container>
  );
};

export default StreamAPI;