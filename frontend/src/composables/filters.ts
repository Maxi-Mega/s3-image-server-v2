import { compareSummaries } from "@/composables/images";
import type { ImageSummary } from "@/models/image";

export function applyFilters(
  summaries: Array<ImageSummary>,
  groupsAndTypes: Record<string, string[]>,
  search: string
): ImageSummary[] {
  let filtered = summaries.filter(
    // @ts-expect-error it can't be undefined
    (img) => img.group in groupsAndTypes && groupsAndTypes[img.group].includes(img.type)
  );

  if (search) {
    search = search.toLowerCase();
    filtered = filtered.filter(
      (img) => img.name.toLowerCase().includes(search) || img.key.toLowerCase().includes(search)
    );
  }

  return filtered.sort(compareSummaries);
}
