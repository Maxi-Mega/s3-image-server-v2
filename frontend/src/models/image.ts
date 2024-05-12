import type { CachedObject } from "@/models/common";
import type { Geonames } from "@/models/geonames";
import type { Localization } from "@/models/localization";
import type { Features } from "@/models/features";

export interface ImageSummary extends CachedObject {
  bucket: string;
  key: string;
  name: string;
  group: string;
  type: string;
  features: Features;

  _hasBeenUpdated: boolean;
}

export interface Image {
  imageSummary: ImageSummary & { cachedObject: CachedObject };
  geonames: Geonames;
  localization: Localization;
  additionalFiles: Record<string, string>;
  targetFiles: Array<string>;
  fullProductFiles: Record<string, string>;

  _lastModified: string;
  _links: Array<[string, string]>; // [filename, URL]
}
