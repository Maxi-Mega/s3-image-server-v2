import { type CachedObject } from "@/models/common";
import { type Geonames } from "@/models/geonames";
import { type Localization } from "@/models/localization";
import type { ProductInformation } from "@/models/product_info.ts";

export type ImageSize = {
  width: number;
  height: number;
};

export class ImageSummary {
  bucket: string;
  key: string;
  name: string;
  group: string;
  type: string;
  geonames: Geonames | null;
  productInfo: ProductInformation | null;
  dynamicFilters: Record<string, string>;
  cachedObject: CachedObject;
  size: ImageSize;

  _hasBeenUpdated: boolean;
  _lastModified: Date;

  constructor(
    bucket: string,
    key: string,
    name: string,
    group: string,
    type: string,
    geonames: Geonames | null,
    productInfo: ProductInformation | null,
    dynamicFilters: Record<string, string>,
    cachedObject: CachedObject,
    size: ImageSize,
    hasBeenUpdated: boolean,
    _lastModified: Date
  ) {
    this.bucket = bucket;
    this.key = key;
    this.name = name;
    this.group = group;
    this.type = type;
    this.geonames = geonames;
    this.productInfo = productInfo;
    this.dynamicFilters = dynamicFilters;
    this.cachedObject = cachedObject;
    this.size = size;
    this._hasBeenUpdated = hasBeenUpdated;
    this._lastModified = _lastModified;
  }
}

export class Image {
  imageSummary: ImageSummary & { cachedObject: CachedObject };
  localization: Localization | null;
  cachedFileLinks: Record<string, string>;
  targetFiles: Array<string>;
  signedURLs: Record<string, string>;

  _lastModified: string;
  _links: Array<[string, string]>; // [filename, URL]

  constructor(
    imageSummary: ImageSummary & {
      cachedObject: CachedObject;
    },
    localization: Localization | null,
    additionalFiles: Record<string, string>,
    targetFiles: Array<string>,
    fileLinks: Record<string, string>,
    lastModified: string,
    links: Array<[string, string]>
  ) {
    this.imageSummary = imageSummary;
    this.localization = localization;
    this.cachedFileLinks = additionalFiles;
    this.targetFiles = targetFiles;
    this.signedURLs = fileLinks;
    this._lastModified = lastModified;
    this._links = links;
  }
}
