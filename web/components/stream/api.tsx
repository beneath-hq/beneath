import { Typography } from "@material-ui/core";
import { makeStyles, Theme } from "@material-ui/core/styles";
import { Table, TableBody, TableCell, TableHead, TableRow } from "components/Tables";
import CodePaper from "components/CodePaper";
import { Link } from "components/Link";
import { GATEWAY_URL } from "lib/connection";
import { toURLName } from "lib/names";
import { FC } from "react";
import useMe from "hooks/useMe";
import { StreamSchemaKind } from "apollo/types/globalTypes";

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
  schemaKind: StreamSchemaKind;
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
          content: buildPythonPipelines(args),
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
  const exampleRecord = makeExamplePythonRecord(args);
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
        WARNING: Using <code>write_full</code> will delete the stream's current primary version and all its data.
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
    await stream.write(${indent(exampleRecord, 4)})


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

const buildPythonPipelines = (args: TemplateArgs) => {
  const exampleRecord = makeExamplePythonRecord(args);
  return (
    <>
      <Heading>Introduction</Heading>
      <Para>
        Pipelines provide an abstraction over the basic Beneath APIs that makes it easier to develop, test, and deploy
        stream processing logic.
      </Para>
      <Para>
        Beneath pipelines are currently quite basic and do not yet support joins and aggregations. They are still
        well-suited for generating streams, one-to-N stream derivation, as well as syncing and alerting records.
      </Para>
      <Heading>Setup</Heading>
      <Para>
        If you haven't already, install the Beneath library and authenticate your environment by{" "}
        <Link href="https://about.beneath.dev/docs/quick-starts/install-sdk/">following this guide</Link>.
      </Para>
      <Heading>Deriving a new stream</Heading>
      <Para>
        The snippet below shows how to create a pipeline that derives data from this stream into a new stream:
      </Para>
      <CodePaper language="python" paragraph filename="pipeline.py">{`
import beneath

async def derive(record):
    result = ... # 1. Derive a new record (can also be a list of records)
    return result

p = beneath.Pipeline(parse_args=True)
t1 = p.read_stream("${args.organization}/${args.project}/${args.stream}")
t2 = p.apply(t1, derive)
p.write_stream(
    t2,
    stream_path="derived-results", # 2. Output stream name
    # 3. Output stream schema
    schema="""
        type Result @schema {
          ...
        }
    """,
)

if __name__ == "__main__":
    p.main()
      `}</CodePaper>
      <Para>To test the pipeline, run:</Para>
      <CodePaper language="bash" paragraph>
        python pipeline.py test
      </CodePaper>
      <Para>See the last section on this page for details on how to run and deploy pipelines.</Para>
      <Heading>Generating records for this stream</Heading>
      <Para>The snippet below shows how to create a pipeline that generates data and writes it to the stream:</Para>
      <CodePaper language="python" paragraph filename="pipeline.py">{`
import asyncio
import beneath
from datetime import datetime, timezone

async def generate_fn(p):
    # The pipeline has a checkpointer, which lets you persist small values.
    # If you're building a scraper, you often use it to track a datetime.
    current_time = await p.checkpoints.get("KEY", default=datetime(1970, 1, 1, tzinfo=timezone.utc))

    # Generators often use a loop to fetch or create the generated records
    while True:
        # Fetch records based on current_time
        data = ...

        # Create and yield generated records
        for row in data:
          yield ${indent(exampleRecord, 12)}

        # Update current_time and checkpoint it
        current_time = ...
        p.checkpoints.set("KEY", current_time)

        # Wait before fetching again or return 
        await asyncio.sleep(1) # or return to stop the pipeline

p = beneath.Pipeline(parse_args=True)
t1 = p.generate(generate_fn)
p.write_stream(
    t1,
    stream_path="${args.organization}/${args.project}/${args.stream}",
    schema="""
${dynamicIndent(args.schema, 8)}
    """,${args.schemaKind === "GraphQL" ? "" : '\n    schema_kind="' + args.schemaKind + '",'}
)

if __name__ == "__main__":
    p.main()

      `}</CodePaper>
      <Para>To test the pipeline, run:</Para>
      <CodePaper language="bash" paragraph>
        python pipeline.py test
      </CodePaper>
      <Para>See the next section on this page for details on how to run and deploy pipelines.</Para>
      <Heading>Running and deploying pipelines</Heading>
      <Para>Running a pipeline requires a couple steps.</Para>
      <Para>
        First, you <em>stage</em> the pipeline in a project to create all the resources that the pipeline relies upon,
        such as output streams, a service and a checkpointer:
      </Para>
      <CodePaper language="bash" paragraph>
        python pipeline.py stage USERNAME/PROJECT/NAME
      </CodePaper>
      <Para>You're now ready to run the pipeline:</Para>
      <CodePaper language="bash" paragraph>
        python pipeline.py run USERNAME/PROJECT/NAME
      </CodePaper>
      <Para>
        The run command has several interesting configurations, such as output versioning and delta processing (where
        the pipelines processes all changes since it last ran, then quits). For more details, run:
      </Para>
      <CodePaper language="bash" paragraph>
        python pipeline.py run -h
      </CodePaper>
      <Para>
        If you're deploying your pipeline to production, you should use a{" "}
        <Link href="https://about.beneath.dev/docs/misc/resources/#services">service</Link>. When staging a pipeline, a
        service is automatically created. To get a secret for it, run:
      </Para>
      <CodePaper language="bash" paragraph>
        beneath service issue-secret USERNAME/PROJECT/NAME --description "My production secret"
      </CodePaper>
      <Para>
        You use the secret by setting the `BENEATH_SECRET` environment variable in your production environment. You can
        also test it locally from the command-line:
      </Para>
      <CodePaper language="bash" paragraph>
        BENEATH_SECRET=... python pipeline.py run USERNAME/PROJECT/NAME
      </CodePaper>
      <Para>
        Navigate to <code>beneath.dev/USERNAME/PROJECT</code> in your browser to get an overview of all the staged
        resources, records flowing in your streams, and usage and monitoring for your service!
      </Para>
      <Para>If you want to remove or reset a staged pipeline completely, run teardown:</Para>
      <CodePaper language="bash" paragraph>
        python pipeline.py teardown USERNAME/PROJECT/NAME
      </CodePaper>
    </>
  );
};

const indent = (code: string, spaces: number) => {
  const spacesString = " ".repeat(spaces);
  return code.replace(/\n/g, "\n" + spacesString);
};

const dynamicIndent = (code: string, spaces: number) => {
  // detect least number of prefix spaces
  let minSpaces = 99;
  const lines = code.match(/^[ \t]*/gm);
  if (lines === null || lines.length === 0) {
    minSpaces = 0;
  } else {
    for (const line of lines) {
      const expanded = line.replace("\t", "    ");
      if (expanded.length < minSpaces) {
        minSpaces = expanded.length;
      }
    }
  }
  const spacesString = " ".repeat(spaces - minSpaces);
  return spacesString + code.replace(/\n/g, "\n" + spacesString);
};

const makeExamplePythonRecord = (args: TemplateArgs) => {
  let result = "{\n";
  const parsedSchema = JSON.parse(args.avroSchema);
  for (const field of parsedSchema["fields"]) {
    result += `    "${field.name}": ...,\n`;
  }
  result += "}";
  return result;
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
