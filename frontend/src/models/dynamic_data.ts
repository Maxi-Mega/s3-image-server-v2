export type FileSelector = {
  regex: string;
  kind: string;
  link: boolean;
};

export type DynamicData = {
  fileSelectors: Record<string, FileSelector>;
  expressions: Record<string, string>;
};
