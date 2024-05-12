import type { CachedObject } from "@/models/common";

export interface Features {
  class: string;
  count: number;
  objects: Record<string, number>;
  cachedObject: CachedObject;
}
