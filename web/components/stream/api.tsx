import { Container, Grid, makeStyles, Tab, Tabs, Theme, Typography } from "@material-ui/core";
import { toURLName } from "lib/names";
import CodePaper from "components/CodePaper";
import { Link } from "components/Link";
import { GATEWAY_URL } from "lib/connection";
import { FC } from "react";

export interface TemplateArgs {
  organization: string;
  project: string;
  stream: string;
  schema: string;
}

export const buildTemplate = (args: TemplateArgs) => {
  args.organization = toURLName(args.organization);
  args.project = toURLName(args.project);
  args.stream = toURLName(args.stream);
  return [
    {
      label: "Python",
      tabs: [
        {
          label: "Reading",
          content: (
            <>
              <Typography paragraph>Install the Beneath Python library:</Typography>
              <CodePaper language="bash" paragraph>{`pip install beneath`}</CodePaper>
              <Typography paragraph>Authenticate your environment:</Typography>
              <CodePaper language="bash" paragraph>{`beneath auth SECRET`}</CodePaper>
              <Typography variant="body1" paragraph>
                From a Python script or notebook:
              </Typography>
              <CodePaper language="python" paragraph>
                {`import beneath

df = await beneath.easy_read("${args.organization}/${args.project}/${args.stream}")`}
              </CodePaper>
            </>
          ),
        },
        {
          label: "Writing",
          content: (
            <>
              <Typography variant="body1" paragraph>
                From a Python script or notebook:
              </Typography>
              <CodePaper language="python" paragraph>
                {`
import beneath
client = beneath.Client()

schema = """
${args.schema}
"""

stream = await client.create_stream(stream_path="${args.organization}/${args.project}/${args.stream}", schema=schema)
instance = await stream.create_instance(version=VERSION)
async with instance.writer() as w:
    await w.write(LIST_OF_RECORDS)`}
              </CodePaper>
            </>
          ),
        },
        {
          label: "Pipelines",
          content: (
            <>
              <Typography paragraph>
                Beneath Pipelines make it easy to do stream processing on any Beneath data stream.
              </Typography>
              <Typography paragraph>
                First, create a new Python file for your pipeline logic. Choose one from the "Consume," "Derive," or
                "Advanced" options below.
              </Typography>
              <Typography variant="subtitle1" gutterBottom>
                Consume: apply a user-defined function
              </Typography>
              <CodePaper language="python" paragraph filename={"your_pipeline.py"}>
                {`import beneath

async def consume_fn(record):
  # YOUR LOGIC HERE

beneath.easy_consume_stream(
  input_stream_path="${args.organization}/${args.project}/${args.stream}",
  consume_fn=consume_fn,
)`}
              </CodePaper>
              <Typography variant="subtitle1" gutterBottom>
                Derive: apply a user-defined function and write results to a new stream
              </Typography>
              <CodePaper language="python" paragraph filename={"your_pipeline.py"}>
                {`import beneath

OUTPUT_SCHEMA="""
YOUR OUTPUT SCHEMA HERE
"""

async def apply_fn(record):
  # YOUR LOGIC HERE
  yield new_record

beneath.easy_derive_stream(
  input_stream_path="${args.organization}/${args.project}/${args.stream}",
  apply_fn=apply_fn,
  output_stream_path="USER/PROJECT/YOUR_NEW_STREAM_NAME",
  output_stream_schema=OUTPUT_SCHEMA
)`}
              </CodePaper>
              <Typography variant="subtitle1" gutterBottom>
                Advanced: check out the full docs to see what you can do with an advanced pipeline
              </Typography>
              <Typography variant="body1" paragraph>
                Second, <Link href={"/-/create/service"}>create a service</Link>, then stage your pipeline:
              </Typography>
              <CodePaper
                language="bash"
                paragraph
              >{`python your_pipeline.py stage USERNAME/PROJECT/SERVICE`}</CodePaper>
              <Typography variant="body1" paragraph>
                Lastly, run your pipeline:
              </Typography>
              <CodePaper language="bash" paragraph>{`python your_pipeline.py run USERNAME/PROJECT/SERVICE`}</CodePaper>
            </>
          ),
        },
      ],
    },
    {
      label: "JavaScript",
      tabs: [
        {
          label: "Reading",
          content: (
            <>
              <Typography variant="body1" paragraph>
                Install the Javascript library with npm:
              </Typography>
              <CodePaper language="bash" paragraph>{`npm install beneath`}</CodePaper>
              <Typography variant="body1" paragraph>
                Or install with yarn:
              </Typography>
              <CodePaper language="bash" paragraph>{`yarn install beneath`}</CodePaper>
              <Typography variant="body1" paragraph>
                You can query this stream directly from your frontend. It's very important to use permissioned
                "read-only" secrets in your frontend code.
              </Typography>
              <CodePaper language={"javascript"}>
                {`fetch("${GATEWAY_URL}/v1/${args.organization}/${args.project}/${args.stream}", {
    "Authorization": "Bearer SECRET",
    "Content-Type": "application/json",
  })
  .then(res => res.json())
  .then(data => {
    // TODO: Add your logic here
    console.log(data)
  })`}
              </CodePaper>
            </>
          ),
        },
        {
          label: "React",
          content: (
            <>
              <Typography variant="body1" paragraph>
                Install the React library with npm:
              </Typography>
              <CodePaper language="bash" paragraph>{`npm install beneath-react`}</CodePaper>

              <Typography variant="body1" paragraph>
                Or install with yarn:
              </Typography>
              <CodePaper language="bash" paragraph>{`yarn install beneath-react`}</CodePaper>
              <Typography variant="body1" paragraph>
                You can query this stream directly from your frontend. It's very important to use permissioned
                "read-only" secrets in your frontend code.
              </Typography>
              <CodePaper language={"javascript"}>
                {`import { useRecords } from "beneath-react";

const App = () => {
  const { records, error, loading } = useRecords({
    stream: "${args.organization}/${args.project}/${args.stream}",
    query: {type: "log", peek: "true"},
    pageSize: 500,
  });
}`}
              </CodePaper>
            </>
          ),
        },
      ],
    },
    {
      label: "SQL",
      tabs: [
        {
          label: "Reading",
          content: (
            <>
              <Typography variant="body1" paragraph>
                You can query this stream in the <Link href="/-/sql">Beneath SQL Editor</Link> like this:
              </Typography>
              <CodePaper
                language={"sql"}
              >{`select * from \`${args.organization}/${args.project}/${args.stream}\``}</CodePaper>
              <Typography variant="body1">
                Additionally, you can connect any business intelligence tool that you'd like, such as Metabase, Tableau,
                Mode, etc. Please reach out if you'd to do this, as it currently takes just a little hand-holding.
              </Typography>
            </>
          ),
        },
      ],
    },
    {
      label: "REST",
      tabs: [
        {
          label: "Reading",
          content: (
            <>
              <Typography variant="body1" paragraph>
                You can query this data stream from anywhere using the REST API. Here's an example using cURL from the
                command line:
              </Typography>
              <CodePaper language="bash" paragraph>
                {`curl -H "Authorization: Bearer SECRET" ${GATEWAY_URL}/v1/${args.organization}/${args.project}/${args.stream}`}
              </CodePaper>
            </>
          ),
        },
      ],
    },
  ];
};
