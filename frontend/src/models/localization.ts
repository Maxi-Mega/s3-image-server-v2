import type { CachedObject } from "@/models/common";

export interface Point {
  coordinates: {
    lon: number;
    lat: number;
  };
}

export interface LocalizationCorner {
  upperLeft: Point;
  upperRight: Point;
  lowerLeft: Point;
  lowerRight: Point;
}

export interface Localization {
  corner: LocalizationCorner;
  cachedObject: CachedObject;
}
