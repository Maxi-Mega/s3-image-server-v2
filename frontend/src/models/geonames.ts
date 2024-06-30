import { type CachedObject } from "@/models/common";

export class GeonamesObject {
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

  constructor(name: string, states: any) {
    this.name = name;
    this.states = states;
  }
}

export class Geonames {
  objects: GeonamesObject[];
  cachedObject: CachedObject;

  constructor(objects: GeonamesObject[], cachedObject: CachedObject) {
    this.objects = objects;
    this.cachedObject = cachedObject;
  }
}
