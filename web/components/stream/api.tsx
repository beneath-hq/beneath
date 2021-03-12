import { Typography } from "@material-ui/core";
import { makeStyles, Theme } from "@material-ui/core/styles";
import { Table, TableBody, TableCell, TableHead, TableRow } from "components/Tables";
import CodePaper from "components/CodePaper";
import { Link } from "components/Link";
import { GATEWAY_URL } from "lib/connection";
import { toURLName } from "lib/names";
import { FC } from "react";
import useMe from "hooks/useMe";

const useStyles = makeStyles((theme: Theme) => ({
  heading: {
    "&:not(:first-child)": {
      marginTop: "2.5rem",
    },
  },
}));

const Heading: FC = (props) => {
  const classes = useStyles();
  return <Typography className={classes.heading} variant="h2" paragraph {...props} />;
};

const Para: FC = (props) => <Typography paragraph {...props} />;

const SecretsLink: FC = (props) => {
  const me = useMe();
  if (!me) {
    return <Link href="/-/auth" {...props} />;
  }
  return (
    <Link
      href={`/organization?organization_name=${toURLName(me.name)}&tab=secrets`}
      as={`/${toURLName(me.name)}/-/secrets`}
      {...props}
    />
  );
};

export interface TemplateArgs {
  organization: string;
  project: string;
  stream: string;
  schema: string;
  avroSchema: string;
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
          content: buildPythonReading(args),
        },
        {
          label: "Writing",
          content: buildPythonWriting(args),
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
      label: "React",
      tabs: [{ label: "Reading", content: buildJavaScriptReact(args) }],
    },
    {
      label: "REST",
      tabs: [
        { label: "Reading", content: buildRESTReading(args) },
        { label: "Writing", content: buildRESTWriting(args) },
      ],
    },
  ];
};

const buildPythonReading = (args: TemplateArgs) => {
  return (
    <>
      <Heading>Setup</Heading>
      <Para>
        If you haven't already, install the Beneath library and authenticate your environment by{" "}
        <Link href="https://about.beneath.dev/docs/quick-starts/install-sdk/">following this guide</Link>.
      </Para>
      <Heading>Read the entire stream into memory</Heading>
      <Para>
        This snippet loads the entire stream into a Pandas DataFrame, which is useful for analysis in notebooks or
        scripts:
      </Para>
      <CodePaper language="python" paragraph>{`
import beneath

df = await beneath.query_index("${args.organization}/${args.project}/${args.stream}")
      `}</CodePaper>
      <Para>
        The function accepts several optional arguments. The most common are <code>to_dataframe=False</code> to get
        records as a regular Python list, <code>filter="..."</code> to{" "}
        <Link href="https://about.beneath.dev/docs/reading-writing-data/index-filters/">filter</Link> by key fields, and{" "}
        <code>max_bytes=...</code> to increase the cap on how many records to load (used to prevent runaway costs). For
        more details, see{" "}
        <Link href="https://python.docs.beneath.dev/easy.html#beneath.easy.query_index">the API reference</Link>.
      </Para>
      <Heading>Replay the stream's history and subscribe to changes</Heading>
      <Para>
        This snippet replays the stream's historical records one-by-one and stays subscribed to new records, which is
        useful for alerting and data enrichment:
      </Para>
      <CodePaper language="python" paragraph>{`
import beneath

async def callback(record):
    print(record)

await beneath.consume("${args.organization}/${args.project}/${args.stream}", callback)
      `}</CodePaper>
      <Para>
        The function accepts several optional arguments. The most common are <code>replay_only=True</code> to stop the
        script once the replay has completed, <code>changes_only=True</code> to only subscribe to changes, and{" "}
        <code>subscription_path="ORGANIZATION/PROJECT/subscription:NAME"</code> to persist the consumer's progress.
      </Para>
      <Heading>Analyze with SQL</Heading>
      <Para>
        This snippet runs a warehouse (OLAP) query on the stream's records and returns the result, which is useful for
        ad-hoc joins, aggregations, and visualizations:
      </Para>
      <CodePaper language="python" paragraph>{`
import beneath

df = await beneath.query_warehouse("SELECT count(*) FROM \`${args.organization}/${args.project}/${args.stream}\`")
      `}</CodePaper>
      <Para>
        See the{" "}
        <Link href="https://about.beneath.dev/docs/reading-writing-data/warehouse-queries/">
          warehouse queries documentation
        </Link>{" "}
        for a guideline to the SQL query syntax.
      </Para>
      <Heading>Reference</Heading>
      <Para>
        Consult the <Link href="https://python.docs.beneath.dev">Beneath Python client API reference</Link> for details
        on all classes, methods and arguments.
      </Para>
    </>
  );
};

const buildPythonWriting = (args: TemplateArgs) => {
  let exampleRecord = "{\n";
  const parsedSchema = JSON.parse(args.avroSchema);
  for (const field of parsedSchema["fields"]) {
    exampleRecord += `    "${field.name}": ...,\n`;
  }
  exampleRecord += "}";

  return (
    <>
      <Heading>Setup</Heading>
      <Para>
        If you haven't already, install the Beneath library and authenticate your environment by{" "}
        <Link href="https://about.beneath.dev/docs/quick-starts/install-sdk/">following this guide</Link>.
      </Para>
      <Heading>Writing basics</Heading>
      <Para>This snippet demonstrates how to connect to the stream and write a record to it:</Para>
      <CodePaper language="python" paragraph>{`
import beneath

client = beneath.Client()
stream = await client.find_stream("${args.organization}/${args.project}/${args.stream}")
await client.start()

await stream.write(${exampleRecord})

await client.stop()
      `}</CodePaper>
      <Para>
        By default, records are buffered in memory for up to one second and sent in batches over the network, allowing
        you to call <code>write</code> many times efficiently (e.g. in a loop). Calling <code>client.stop()</code>{" "}
        ensures all records have been transmitted to Beneath before terminating.
      </Para>
      <Heading>Write an entire dataset in one go</Heading>
      <Para>
        The convenience function <code>write_full</code> allows you to write a full dataset to the stream in one go.
        Each call will create a new version for the stream, and{" "}
        <Link href="https://about.beneath.dev/docs/concepts/streams/#streams-can-have-multiple-versions-known-as-_instances_">
          finalize
        </Link>{" "}
        it once the writes have completed.
      </Para>
      <Para>
        WARNING: Using <code>write_full</code> will delete the current version of the stream and all its data.
      </Para>
      <CodePaper language="python" paragraph>{`
import beneath

df = pd.DataFrame(...)
await beneath.write_full("${args.organization}/${args.project}/${args.stream}", df, recreate_on_schema_change=True)
`}</CodePaper>
      <Heading>Writing records from a web server</Heading>
      <Para>
        A frequent use case for Beneath is to create an API that writes data to a stream. This example shows how to do
        so using <Link href="https://fastapi.tiangolo.com">FastAPI</Link>, which is like Flask, but faster and with
        better support for <code>async</code> and <code>await</code>.
      </Para>
      <Para>First install the dependencies:</Para>
      <CodePaper language="bash" paragraph>
        pip install fastapi uvicorn
      </CodePaper>
      <Para>
        Then create the web server (edit the <code>post</code> function or add your own endpoints):
      </Para>
      <CodePaper language="python" paragraph filename="server.py">{`
import beneath
import uvicorn
from fastapi import FastAPI

app = FastAPI()

client = beneath.Client()
stream = None


@app.on_event("startup")
async def on_startup():
    global stream
    stream = await client.find_stream("${args.organization}/${args.project}/${args.stream}")
    await client.start()


@app.on_event("shutdown")
async def on_shutdown():
    await stream.stop()


@app.post("/")
async def post(payload: dict):
    # TODO: Validate and use payload
    # NOTE: Don't write payload directly unless you trust the user
    await stream.write(${exampleRecord})


if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
`}</CodePaper>
      <Para>Run it from the command-line:</Para>
      <CodePaper language="bash" paragraph>
        uvicorn server:app --reload
      </CodePaper>
      <Para>Test the API using cURL:</Para>
      <CodePaper language="bash" paragraph>{`
curl http://localhost:8000 \\
  -X POST \\
  -H "Content-Type: application/json" \\
  -d 'PAYLOAD'
`}</CodePaper>
    </>
  );
};

const buildJavaScriptReact = (args: TemplateArgs) => {
  return (
    <>
      <Heading>Setup</Heading>
      <Para>First, add the Beneath react client to your project:</Para>
      <CodePaper language="bash" paragraph>{`npm install beneath-react`}</CodePaper>
      <Para>
        To query private streams, create a read-only secret on your <SecretsLink>secrets</SecretsLink> page. If you're
        going to use it in production, create a{" "}
        <Link href="https://about.beneath.dev/docs/reading-writing-data/access-management/#creating-services-setting-quotas-and-granting-permissions">
          service secret
        </Link>
        .
      </Para>
      <Heading>Reading</Heading>
      <Para>
        The <code>useRecords</code> hook allows you to run log and index queries, paginate results, and subscribe to
        changes with websockets. Use the example below to get started:
      </Para>
      <CodePaper language="jsx">
        {`
import { useRecords } from "beneath-react";

const App = () => {
  const { records, loading, error } = useRecords({
    stream: "${args.organization}/${args.project}/${args.stream}",
    // Other useful options:
    // secret: "INSERT",
    // query: { type: "log", peek: false },
    // query: { type: "index", filter: { ... } },
    // subscribe: true,
  })

  if (loading) {
    return <p>Loading...</p>;
  } else if (error) {
    return <p>Error: {error}</p>;
  }

  return (
    <div>
      <h1>${args.stream}</h1>
      <ul>
        {records.map((record) => (
          <li key={record["@meta"].key}>
            {JSON.stringify(record)}
          </li>
        ))}
      </ul>
    </div>
  );
}
        `}
      </CodePaper>
      <Heading>Reference</Heading>
      <Para>
        Consult the <Link href="https://react.docs.beneath.dev">Beneath React client API reference</Link> for details.
        There's also a lower-level <Link href="https://js.docs.beneath.dev">vanilla JavaScript client</Link>.
      </Para>
    </>
  );
};

const buildRESTReading = (args: TemplateArgs) => {
  const url = `${GATEWAY_URL}/v1/${args.organization}/${args.project}/${args.stream}`;
  return (
    <>
      <Heading>Reading basics</Heading>
      <Para>
        Query the stream <strong>log</strong> with cURL:
      </Para>
      <CodePaper language="bash" paragraph>
        {`
curl ${url} \\
  -H "Authorization: Bearer SECRET" \\
  -d type=log \\
  -d limit=25 \\
  -G
`}
      </CodePaper>
      <Para>
        Query the stream <strong>index</strong> with cURL (the filter is optional):
      </Para>
      <CodePaper language="bash" paragraph>
        {`
curl ${url} \\
  -H "Authorization: Bearer SECRET" \\
  -d type=index \\
  -d filter=FILTER \\
  -d limit=25 \\
  -G
`}
      </CodePaper>
      <Para>Sample response:</Para>
      <CodePaper language="json" paragraph>{`
{
    "data": [
        {
            "@meta": {
                "key": "AkhlbGxvABYH4w==",
                "timestamp": 1612947705590
            },
            ....
        },
        ...
    ],
    "meta": {
        "instance_id": "e144393c-6fd8-4e56-b9e5-e994c90f4bda",
        "next_cursor": "3Fd28XgzvydG755KnfRQugHqiAzoH61",
        "change_cursor": "7ht7w4GiRcQztEEdzEG8gif4g4yX",
    }
}
      `}</CodePaper>
      <Heading>Cursors (pagination and changes)</Heading>
      <Para>
        You can use <code>next_cursor</code> to get the next page of results and <code>change_cursor</code> to
        fetch/poll for new records in the stream's log. To fetch records with a cursor:
      </Para>
      <CodePaper language="bash" paragraph>
        {`
curl ${url} \\
  -H "Authorization: Bearer SECRET" \\
  -d cursor=CURSOR \\
  -d limit=25 \\
  -G
`}
      </CodePaper>
      <Para>
        When you use <code>change_cursor</code>, the response will place the next change cursor in the{" "}
        <code>next_cursor</code> field.
      </Para>
      <Para>
        Note: The <code>change_cursor</code> returned from filtered index queries will match all new records, not just
        the filter.
      </Para>
      <Heading>Parameters</Heading>
      <Para>Here follows a full list of parameters for reading data from streams over REST.</Para>
      <Table>
        <TableHead>
          <TableRow>
            <TableCell>Parameter</TableCell>
            <TableCell>Value</TableCell>
            <TableCell>Description</TableCell>
          </TableRow>
        </TableHead>
        <TableBody>
          <TableRow>
            <TableCell>
              <code>limit</code>
            </TableCell>
            <TableCell>integer</TableCell>
            <TableCell>The maximum number of records to return per page</TableCell>
          </TableRow>
          <TableRow>
            <TableCell>
              <code>cursor</code>
            </TableCell>
            <TableCell>string</TableCell>
            <TableCell>
              A cursor to fetch more records from. Cursors are returned from paginated log or index requests.
            </TableCell>
          </TableRow>
          <TableRow>
            <TableCell>
              <code>type</code>
            </TableCell>
            <TableCell>"log" or "index"</TableCell>
            <TableCell>
              The type of query to run. "log" queries return all records in write order. "index" queries return the most
              recent record for each key, sorted and optionally filtered by key. See the{" "}
              <Link href="https://about.beneath.dev/docs/concepts/unified-data-system/">docs</Link> for more.
            </TableCell>
          </TableRow>
          <TableRow>
            <TableCell>
              <code>peek</code>
            </TableCell>
            <TableCell>"true" or "false"</TableCell>
            <TableCell>
              Only applies if <code>type=log</code>. If false, returns log records from oldest to newest. If true,
              returns from newest to oldest.
            </TableCell>
          </TableRow>
          <TableRow>
            <TableCell>
              <code>filter</code>
            </TableCell>
            <TableCell>JSON</TableCell>
            <TableCell>
              Only applies if type=index. If set, the filter is applied to the index and only matching record(s)
              returned. See the{" "}
              <Link href="https://about.beneath.dev/docs/reading-writing-data/index-filters/">filter docs</Link> for
              syntax.
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </>
  );
};

const buildRESTWriting = (args: TemplateArgs) => {
  const url = `${GATEWAY_URL}/v1/${args.organization}/${args.project}/${args.stream}`;
  return (
    <>
      <Heading>Writing basics</Heading>
      <Para>Write records with cURL:</Para>
      <CodePaper language="bash" paragraph>
        {`
curl ${url} \\
  -X POST \\
  -H "Authorization: Bearer SECRET" \\
  -H "Content-Type: application/json" \\
  -d '[{"field": "value", ...}, ...]'
`}
      </CodePaper>
      <Para>
        Replace the last line with the JSON-encoded record(s) you're writing. A succesful write returns HTTP status code
        200. You should see the records appear in the stream shortly after.
      </Para>
      <Heading>Encoding records as JSON</Heading>
      <Para>
        Beneath's schemas support more data types than JSON does. The{" "}
        <Link href="https://about.beneath.dev/docs/reading-writing-data/index-filters/#representing-data-types-as-json">
          Representing data types as JSON
        </Link>{" "}
        section of the filtering documentation shows how to encode different data types as JSON.
      </Para>
    </>
  );
};
