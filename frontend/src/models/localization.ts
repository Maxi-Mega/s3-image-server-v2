import type { CachedObject } from "@/models/common";

export interface Point {
  coordinates: {
    lon: number;
    lat: number;
  };
}

export interface LocalizationCorner {
  "upper-left": Point;
  "upper-right": Point;
  "lower-left": Point;
  "lower-right": Point;
}

export interface Localization {
  corner: LocalizationCorner;
  cachedObject: CachedObject;
}
