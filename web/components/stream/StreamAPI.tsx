import React, { FC, useState } from "react";
import { Alert, TabContext, TabPanel, ToggleButton, ToggleButtonGroup } from "@material-ui/lab";
import { AppBar, Container, Grid, Link as MUILink, makeStyles, Tab, Tabs, Theme, Typography, useMediaQuery, useTheme } from "@material-ui/core";
import CodeIcon from '@material-ui/icons/Code';

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
  const me = useMe();
  const classes = useStyles();
  const [language, setLanguage] = useState('Python');
  const [pythonDetail, setPythonDetail] = useState('Reading');
  const [javascriptDetail, setJavascriptDetail] = useState('Reading');
  const [reactDetail, setReactDetail] = useState('Reading');

  const languageTabs = ["Python", "Javascript", "React", "SQL", "cURL"];
  const pythonTabs = ["Setup", "Reading", "Writing", "Pipelines"];
  const javascriptTabs = ["Setup", "Reading"];
  const reactTabs = ["Setup", "Reading"];

  return (
    <Container maxWidth="md" className={classes.container}>
      <Alert severity="info">
        {me && (
          <>
            To create a secret for connecting to Beneath, head to your{" "}
            <Link
              href={`/organization?organization_name=${toURLName(me.name)}&tab=secrets`}
              as={`/${toURLName(me.name)}/-/secrets`}
            >
              profile page
            </Link>
          </>
        )}
        {!me && (
          <>
            You'll first have to <Link href="/-/auth">create a user</Link> to get a secret for
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
client = beneath.Client()
df = await client.easy_read("${toURLName(stream.project.organization.name)}/${toURLName(stream.project.name)}/${toURLName(stream.name)}")`}
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

stream = await client.stage_stream(stream_path="${toURLName(stream.project.organization.name)}/${toURLName(stream.project.name)}/${toURLName(stream.name)}", schema=schema)
instance = await stream.stage_instance(version=VERSION)
async with instance.writer() as w:
    await w.write(LIST_OF_RECORDS)`}
              </CodeBlock>
            </>
          )}
          {language === "Python" && pythonDetail === "Pipelines" && (
            <>
              <Typography variant="body1" paragraph>
                Create a new Python file with your pipeline logic:
              </Typography>
              <CodeBlock language={"python"} title={"your_pipeline.py"}>
{`import beneath
p = beneath.Pipeline(parse_args=True)

async def fn(record):
  # YOUR LOGIC HERE

data = p.read_stream("${toURLName(stream.project.organization.name)}/${toURLName(stream.project.name)}/${toURLName(stream.name)}")
new_data = p.apply(data, fn)
p.write_stream(new_data, "OUTPUT_STREAM_PATH") # optional
p.main()`}
              </CodeBlock>
              <VSpace units={2} />
              <Typography variant="body1" paragraph>
                <Link href={"/-/create/service"}>Create a service</Link>, then stage your pipeline:
              </Typography>
              <CodeBlock language={"bash"}>
{`python your_pipeline.py stage USERNAME/PROJECT/SERVICE`}
              </CodeBlock>
              <VSpace units={2} />
              <Typography variant="body1" paragraph>
                Run your pipeline:
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
                You can query this stream directly from your front-end. Just copy and paste this snippet to get started. It's
                very important that you only use read-only secrets in your front-end.
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
                You can query this stream directly from your front-end. Just copy and paste this snippet to get started. It's
                very important that you only use read-only secrets in your front-end.
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
                You can query this stream in the <Link href="/-/sql">Beneath SQL Editor</Link> like so:
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
                You can lookup records in this stream from anywhere using the REST API. Here's an example of how to query it
                from the command line:
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