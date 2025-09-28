import { resolveBackendURL } from "@/composables/url";
import type { CachedObject } from "@/models/common";
import type { Geonames } from "@/models/geonames";
import { Image, ImageSummary } from "@/models/image";
import type { StaticInfo } from "@/models/static_info";
import { plainToInstance } from "class-transformer";

export type GqlAllSummaries = Record<string, Record<string, ImageSummary[]>>;
export type GqlImage = Record<"getImage", Image>;

export function compareSummaries(a: ImageSummary, b: ImageSummary): number {
  return b._lastModified.getTime() - a._lastModified.getTime();
}

export function processSummaries(summaries: GqlAllSummaries): ImageSummary[] {
  const flattened: ImageSummary[] = [];

  for (const group in summaries) {
    for (const type in summaries[group]) {
      // @ts-expect-error no worries
      flattened.push(...summaries[group][type].map((s) => plainToInstance(ImageSummary, s)));
    }
  }

  return flattened.map((s) => {
    s.cachedObject.lastModified = new Date(s.cachedObject.lastModified);
    return { ...s, _hasBeenUpdated: false, _lastModified: s.cachedObject.lastModified };
  });
}

export function summaryKey(summary: ImageSummary): string {
  return `${summary.bucket}_${summary.key}`;
}

const defaultMaxImagesDisplayCount = 20;

export function limitDisplayedImages(
  summaries: ImageSummary[],
  staticInfo: StaticInfo
): ImageSummary[] {
  return summaries.slice(
    0,
    Math.min(summaries.length, staticInfo.maxImagesDisplayCount ?? defaultMaxImagesDisplayCount)
  );
}

export const wbr = (name: string): string => name.replace(/_/g, "<wbr>_");

export function base(path: string): string {
  path = path.slice(path.lastIndexOf("/") + 1);
  return path.slice(path.lastIndexOf("@") + 1);
}

function fromCachedObj(cachedObject: CachedObject): [string, string] {
  return [base(cachedObject.cacheKey), resolveBackendURL("/api/cache/" + cachedObject.cacheKey)];
}

function makeLinks(img: Image): Array<[string, string]> {
  const links = [] as Array<[string, string]>;

  for (const key in img.fullProductFiles) {
    // The URL of full product files already has its own host
    // @ts-expect-error no worries
    links.push([base(key), img.fullProductFiles[key]]);
  }

  for (const key in img.additionalFiles) {
    links.push([base(key), resolveBackendURL("/api/cache/" + img.additionalFiles[key])]);
  }

  if (img.geonames) links.push(fromCachedObj(img.geonames.cachedObject));
  if (img.localization) links.push(fromCachedObj(img.localization.cachedObject));
  if (img.imageSummary.features) links.push(fromCachedObj(img.imageSummary.features.cachedObject));

  return links;
}

export function formatDate(d: Date): string {
  return d.toISOString().replace("T", " ");
}

export function processImage(image: GqlImage): Image {
  const img = plainToInstance(Image, image.getImage);
  const imageSummary = { ...img.imageSummary, _hasBeenUpdated: false };
  imageSummary.cachedObject.lastModified = new Date(imageSummary.cachedObject.lastModified);
  const _lastModified = formatDate(imageSummary.cachedObject.lastModified);
  const _links = makeLinks(img);
  return { ...img, imageSummary, _lastModified, _links };
}

export function formatGeonames(geonames: Geonames | null): string {
  let final = "";

  if (geonames && geonames.objects) {
    // @ts-expect-error no worries
    final += geonames.objects[0].name;
    // @ts-expect-error no worries
    if (geonames.objects[0].states) {
      // @ts-expect-error no worries
      final += " / " + geonames.objects[0].states[0].name;
      // @ts-expect-error no worries
      if (geonames.objects[0].states[0].counties) {
        // @ts-expect-error no worries
        final += " / " + geonames.objects[0].states[0].counties[0].name;
      }
    }
  }

  return final;
}
