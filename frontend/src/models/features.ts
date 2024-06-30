import type { CachedObject } from "@/models/common";

export class Features {
  class: string;
  count: number;
  objects: Record<string, number>;
  cachedObject: CachedObject;

  constructor(
    clazz: string,
    count: number,
    objects: Record<string, number>,
    cachedObject: CachedObject
  ) {
    this.class = clazz;
    this.count = count;
    this.objects = objects;
    this.cachedObject = cachedObject;
  }
}
