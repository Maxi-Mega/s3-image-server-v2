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
}

export interface Image {
  imageSummary: ImageSummary;
  geonames: Geonames;
  localization: Localization;
  features: Features;
  additionalFiles: Map<string, string>;
  targetFiles: Map<string, string>;
  fullProductFiles: Map<string, string>;
}
