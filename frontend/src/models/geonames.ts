import type { CachedObject } from "@/models/common";

export interface GeonamesObject {
  name: string;
  states: {
    name: string;
    counties: {
      name: string;
      cities: {
        name: string;
      }[];
      villages: {
        name: string;
      }[];
    }[];
  }[];
}

export interface Geonames {
  objects: GeonamesObject[];
  cachedObject: CachedObject;
}
