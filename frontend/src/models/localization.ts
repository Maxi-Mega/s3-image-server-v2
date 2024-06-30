import { type CachedObject } from "@/models/common";

export class Point {
  coordinates: {
    lon: number;
    lat: number;
  };

  constructor(coordinates: { lon: number; lat: number }) {
    this.coordinates = coordinates;
  }
}

export class LocalizationCorner {
  "upper-left": Point;
  "upper-right": Point;
  "lower-left": Point;
  "lower-right": Point;

  constructor(upper_left: Point, upper_right: Point, lower_left: Point, lower_right: Point) {
    this["upper-left"] = upper_left;
    this["upper-right"] = upper_right;
    this["lower-left"] = lower_left;
    this["lower-right"] = lower_right;
  }
}

export class Localization {
  corner: LocalizationCorner;
  cachedObject: CachedObject;

  constructor(corner: LocalizationCorner, cachedObject: CachedObject) {
    this.corner = corner;
    this.cachedObject = cachedObject;
  }
}
