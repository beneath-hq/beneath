import fetch from "isomorphic-unfetch";
import fs from "fs";
import os from "os";

import { BrowserClient } from "./BrowserClient";
import { BENEATH_CONTROL_HOST, DEV } from "./config";
import { StreamQualifier } from "./shared";
import { BrowserJob } from "./BrowserJob";

const PROJECT_NAME = "js_test";
const STREAM_NAME = "foo";
const STREAM_SCHEMA = `
  type Foo @stream @key(fields: "a") {
    a: Int!
    b: String
  }
`;

let client: BrowserClient;
let streamQualifier: { organization: string, project: string, stream: string };

beforeAll(() => {
  const secret = loadLocalSecret();
  client = new BrowserClient({ secret });
});

const makeFoo = (i: number) => {
  return { a: i, b: `The Lord ${i} of Integer` };
};

const loadLocalSecret = () => {
  const secretPath = `${os.homedir()}/.beneath/${DEV ? "secret_dev" : "secret"}.txt`;
  if (fs.existsSync(secretPath)) {
    const secret = fs.readFileSync(secretPath, "utf8");
    if (secret) {
      return secret;
    }
  }
  throw Error("Cannot run tests: local secret not found");
};

const queryControl = async (query: string, variables?: { [key: string]: any; }) => {
  const url = `${BENEATH_CONTROL_HOST}/graphql`;
  const headers: any = { Authorization: `Bearer ${client.secret}`, "Content-Type": "application/json" };
  const body = { query, variables };
  const res = await fetch(url, {
    method: "POST",
    headers,
    body: JSON.stringify(body),
  });
  const json = await res.json();
  if (json.errors || !json.data) {
    throw Error(`Control query failed, response: ${JSON.stringify(json)}`);
  }
  return json.data;
};

test("runs with authenticated CLI and BENEATH_ENV=dev", async () => {
  expect(process.env.BENEATH_ENV).toBe("dev");
  const pong = await client.ping();
  expect(pong.data?.authenticated).toBe(true);
  expect(pong.data?.versionStatus).toBe("stable");
});

test("creates test stream", async () => {
  const meRes = await queryControl("query Me { me { name } }");
  const me = meRes.me;
  expect(me.name).toBeTruthy();

  const projectRes = await queryControl(`
    mutation StageProject($organization: String!, $project: String!) {
			stageProject(organizationName: $organization, projectName: $project) {
				projectID
				name
			}
		}
  `, { organization: me.name, project: PROJECT_NAME });
  const project = projectRes.stageProject;
  expect(project.name).toBe(PROJECT_NAME);

  const streamRes = await queryControl(`
    mutation StageStream($organization: String!, $project: String!, $stream: String!, $schema: String!) {
			stageStream(
				organizationName: $organization,
				projectName: $project,
				streamName: $stream,
				schemaKind: GraphQL,
				schema: $schema,
			) {
				streamID
				name
			}
		}
  `, { organization: me.name, project: project.name, stream: STREAM_NAME, schema: STREAM_SCHEMA });
  const stream = streamRes.stageStream;
  expect(stream.name).toBe(STREAM_NAME);

  const instanceRes = await queryControl(`
    mutation stageStreamInstance($streamID: UUID!) {
			stageStreamInstance(streamID: $streamID, version: 0, makePrimary: true) {
        streamInstanceID
        streamID
			}
		}
  `, { streamID: stream.streamID });
  const instance = instanceRes.stageStreamInstance;
  expect(instance.streamID).toBe(stream.streamID);

  streamQualifier = {
    organization: me.name,
    project: project.name,
    stream: stream.name,
  };
});

test("writes to test stream", async () => {
  const records = [];
  for (let i = 0; i < 50; i++) {
    records.push(makeFoo(i));
  }

  const stream = client.findStream(streamQualifier);
  const { writeID, error } = await stream.write(records);
  expect(error).toBeUndefined();
  expect(writeID).toBeTruthy();
});

test("runs warehouse job and reads results", async () => {
  const query = `select * from \`${streamQualifier.organization}/${streamQualifier.project}/${streamQualifier.stream}\``;
  const { job, error } = await client.queryWarehouse({ query, dry: true });
  expect(error).toBeUndefined();
  expect(job).toBeInstanceOf(BrowserJob);
});
