import { type CachedObject } from "@/models/common";
import { type Geonames } from "@/models/geonames";
import { type Localization } from "@/models/localization";
import { Features } from "@/models/features";

export class ImageSummary {
  bucket: string;
  key: string;
  name: string;
  group: string;
  type: string;
  features: Features | null;
  cachedObject: CachedObject;

  _hasBeenUpdated: boolean;
  _lastModified: Date;

  constructor(
    bucket: string,
    key: string,
    name: string,
    group: string,
    type: string,
    features: Features | null,
    cachedObject: CachedObject,
    hasBeenUpdated: boolean,
    _lastModified: Date
  ) {
    this.bucket = bucket;
    this.key = key;
    this.name = name;
    this.group = group;
    this.type = type;
    this.features = features;
    this.cachedObject = cachedObject;
    this._hasBeenUpdated = hasBeenUpdated;
    this._lastModified = _lastModified;
  }
}

export class Image {
  imageSummary: ImageSummary & { cachedObject: CachedObject };
  geonames: Geonames | null;
  localization: Localization | null;
  additionalFiles: Record<string, string>;
  targetFiles: Array<string>;
  fullProductFiles: Record<string, string>;

  _lastModified: string;
  _links: Array<[string, string]>; // [filename, URL]

  constructor(
    imageSummary: ImageSummary & {
      cachedObject: CachedObject;
    },
    geonames: Geonames | null,
    localization: Localization | null,
    additionalFiles: Record<string, string>,
    targetFiles: Array<string>,
    fullProductFiles: Record<string, string>,
    lastModified: string,
    links: Array<[string, string]>
  ) {
    this.imageSummary = imageSummary;
    this.geonames = geonames;
    this.localization = localization;
    this.additionalFiles = additionalFiles;
    this.targetFiles = targetFiles;
    this.fullProductFiles = fullProductFiles;
    this._lastModified = lastModified;
    this._links = links;
  }
}
