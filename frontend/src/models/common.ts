export class CachedObject {
  lastModified: Date;
  cacheKey: string;

  constructor(lastModified: Date, cacheKey: string) {
    this.lastModified = lastModified;
    this.cacheKey = cacheKey;
  }
}
