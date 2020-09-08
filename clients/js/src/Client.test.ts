import fetch from "isomorphic-unfetch";
import fs from "fs";
import os from "os";

import { BENEATH_CONTROL_HOST, DEV } from "./config";
import { Client } from "./Client";
import { Job } from "./Job";

const PROJECT_NAME = "js_test";
const STREAM_NAME = "foo";
const STREAM_SCHEMA = `
  type Foo @stream @key(fields: "a") {
    a: Int!
    b: String
  }
`;
const NUMROWS = 50;

let client: Client;
let streamQualifier: { organization: string, project: string, stream: string };

beforeAll(() => {
  const secret = loadLocalSecret();
  client = new Client({ secret });
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
  for (let i = 0; i < NUMROWS; i++) {
    records.push(makeFoo(i));
  }

  const stream = client.findStream(streamQualifier);
  const { writeID, error } = await stream.write(records);
  expect(error).toBeUndefined();
  expect(writeID).toBeTruthy();
});

test("runs warehouse job and reads results", async () => {
  jest.setTimeout(30000);

  const query = `
    select a, count(*) as count
    from \`${streamQualifier.organization}/${streamQualifier.project}/${streamQualifier.stream}\`
    group by a
    order by a
  `;

  // dry
  const { job: dryJob, error: dryError } = await client.queryWarehouse({ query, dry: true });
  expect(dryError).toBeUndefined();
  expect(dryJob).toBeInstanceOf(Job);
  expect(dryJob?.jobID).toBeUndefined();
  expect(dryJob?.resultAvroSchema).toBeTruthy();
  expect(dryJob?.referencedInstanceIDs).toHaveLength(1);
  expect(dryJob?.status).toBe("done");
  await expect(dryJob?.getCursor()).rejects.toThrow("Cannot poll dry run job");

  // wet
  const { job, error } = await client.queryWarehouse<{ a: number, count: number }>({ query });
  expect(error).toBeUndefined();
  expect(job).toBeInstanceOf(Job);
  expect(job?.jobID).toBeTruthy();
  expect(job?.status).toBe("running");

  if (!job) { // for satisfying typescript
    fail("job is undefined");
  }

  const { cursor, error: error2 } = await job?.getCursor();
  expect(error2).toBeFalsy();
  expect(cursor).toBeTruthy();
  expect(cursor?.nextCursor).toBeTruthy();
  expect(cursor?.changeCursor).toBeUndefined();
  expect(() => cursor?.subscribeChanges({ onData: () => undefined, onComplete: () => undefined })).toThrowError("cannot subscribe to changes for this query");

  const n = NUMROWS / 2;
  for (let j = 0; j <= NUMROWS; j += n) {
    if (j === NUMROWS) {
      expect(cursor?.hasNext()).toBe(false);
      break;
    }

    const res = await cursor?.readNext({ pageSize: n });
    expect(res?.error).toBeUndefined();

    if (!res?.data) { // for satisfying typescript
      fail("data is undefined");
    }

    expect(res?.data).toHaveLength(n);
    for (let i = 0; i < n; i++) {
      expect(res?.data[i].a).toBe(j + i);
      expect(res?.data[i].count).toBeGreaterThan(0);
    }
  }

  expect(job?.resultAvroSchema).toBeTruthy();
  expect(job?.referencedInstanceIDs).toHaveLength(1);
});
