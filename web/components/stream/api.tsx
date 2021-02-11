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
              <Para>Install the Beneath Python library:</Para>
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
        A succesful write returns HTTP status code 200. You should see the records appear in the stream shortly after.
      </Para>
      <Heading>Encoding records as JSON</Heading>
      <Para>
        Beneath's schemas support more data types than JSON. The{" "}
        <Link href="https://about.beneath.dev/docs/reading-writing-data/index-filters/#representing-data-types-as-json">
          Representing data types as JSON
        </Link>{" "}
        section of the filtering documentation shows how to encode different data types as JSON.
      </Para>
    </>
  );
};
