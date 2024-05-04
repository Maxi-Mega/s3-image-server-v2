import type { Image, ImageSummary } from "@/models/image";

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

export function processImage(image: GqlImage): Image {
  console.info("Image:", image.getImage);

  return image.getImage;
}
