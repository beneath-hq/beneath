export const toURLName = (name: string) => {
  return name.replace("_", "-");
};

export const toBackendName = (name: string) => {
  return name.replace("-", "_");
};
