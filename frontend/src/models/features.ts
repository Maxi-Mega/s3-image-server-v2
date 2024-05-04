import type { CachedObject } from "@/models/common";

export interface Features {
  class: string;
  count: number;
  objects: Map<string, number>;
  cachedObject: CachedObject;
}
