export const toURLName = (name: string) => {
  return name.replace(/_/g, "-");
};

export const toBackendName = (name: string) => {
  return name.replace(/-/g, "_");
};
