import { toURLName } from "lib/names";

interface Stream {
  name: string;
  project: {
    name: string;
    organization: { name: string; };
  };
  primaryStreamInstance?: Instance | null;
}

interface Instance {
  version: number;
  streamInstanceID?: string;
}

export const makeStreamHref = (stream: Stream, instance?: Instance, tab?: string) => {
  let res = `/stream`;
  res += `?organization_name=${toURLName(stream.project.organization.name)}`;
  res += `&project_name=${toURLName(stream.project.name)}`;
  res += `&stream_name=${toURLName(stream.name)}`;
  if (instance && stream.primaryStreamInstance?.version !== instance.version) {
    res += `&version=${instance.version}`;
  }
  if (tab) {
    res += `&tab=${tab}`;
  }
  return res;
};

export const makeStreamAs = (stream: Stream, instance?: Instance, tab?: string) => {
  let res = `/${toURLName(stream.project.organization.name)}`;
  res += `/${toURLName(stream.project.name)}`;
  res += `/stream:${toURLName(stream.name)}`;
  if (instance && stream.primaryStreamInstance?.version !== instance.version) {
    res += `/${instance.version}`;
  }
  if (tab) {
    res += `/-/${tab}`;
  }
  return res;
};
