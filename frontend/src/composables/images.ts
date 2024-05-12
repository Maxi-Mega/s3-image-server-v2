import type { Image, ImageSummary } from "@/models/image";
import type { Geonames } from "@/models/geonames";
import type { CachedObject } from "@/models/common";
import { resolveBackendURL } from "@/composables/url";

export type GqlAllSummaries = Record<string, Record<string, ImageSummary[]>>;
export type GqlImage = Record<"getImage", Image>;

export function compareSummaries(a: ImageSummary, b: ImageSummary): number {
  return b.lastModified.getTime() - a.lastModified.getTime();
}

export function processSummaries(summaries: GqlAllSummaries): ImageSummary[] {
  const flattened: ImageSummary[] = [];

  for (const group in summaries) {
    for (const type in summaries[group]) {
      flattened.push(...summaries[group][type]);
    }
  }

  return flattened.map((s) => {
    const lastModified = new Date(s.lastModified);
    return { ...s, lastModified };
  });
}

let keyCount = 0; // TODO: remove when summaries are no longer duplicated

export function summaryKey(summary: ImageSummary): string {
  return `${summary.bucket} ${summary.group} ${summary.type} ${summary.name} (${keyCount++})`;
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

  if (img.geonames) links.push(fromCachedObj(img.geonames.cachedObject));
  if (img.localization) links.push(fromCachedObj(img.localization.cachedObject));
  if (img.imageSummary.features) links.push(fromCachedObj(img.imageSummary.features.cachedObject));

  for (const key in img.additionalFiles) {
    links.push([base(key), resolveBackendURL("/api/cache/" + img.additionalFiles[key])]);
  }

  for (const key in img.fullProductFiles) {
    // The URL of full product files already has its own host
    links.push([base(key), img.fullProductFiles[key]]);
  }

  return links;
}

export function processImage(image: GqlImage): Image {
  const img = image.getImage;
  const _lastModified = new Date(img.imageSummary.cachedObject.lastModified)
    .toISOString()
    .replace("T", " ");
  const _links = makeLinks(img);
  const targetFiles = img.targetFiles.flatMap((e) => [e, e, e, e, e]); // TODO: remove
  return { ...img, _lastModified, _links, targetFiles };
}

export function formatGeonames(geonames: Geonames | null): string {
  let final = "";

  if (geonames && geonames.objects) {
    final += geonames.objects[0].name;
    if (geonames.objects[0].states) {
      final += " / " + geonames.objects[0].states[0].name;
      if (geonames.objects[0].states[0].counties) {
        final += " / " + geonames.objects[0].states[0].counties[0].name;
      }
    }
  }

  return final;
}
