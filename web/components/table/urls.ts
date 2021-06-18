import { toURLName } from "lib/names";

export interface Table {
  tableID: string;
  name: string;
  project: {
    name: string;
    organization: { name: string; };
  };
  primaryTableInstance?: Instance | null;
}

export interface Instance {
  version: number;
  tableInstanceID?: string;
}

export const makeTableHref = (table: Table, instance?: Instance, tab?: string) => {
  let res = `/table`;
  res += `?organization_name=${toURLName(table.project.organization.name)}`;
  res += `&project_name=${toURLName(table.project.name)}`;
  res += `&table_name=${toURLName(table.name)}`;
  if (instance && table.primaryTableInstance?.version !== instance.version) {
    res += `&version=${instance.version}`;
  }
  if (tab) {
    res += `&tab=${tab}`;
  }
  return res;
};

export const makeTableAs = (table: Table, instance?: Instance, tab?: string) => {
  let res = `/${toURLName(table.project.organization.name)}`;
  res += `/${toURLName(table.project.name)}`;
  res += `/table:${toURLName(table.name)}`;
  if (instance && table.primaryTableInstance?.version !== instance.version) {
    res += `/${instance.version}`;
  }
  if (tab) {
    res += `/-/${tab}`;
  }
  return res;
};
