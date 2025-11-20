import { compareSummaries } from "@/composables/images";
import type { ImageSummary } from "@/models/image";

export function applyFilters(
  summaries: Array<ImageSummary>,
  filters: Record<string, Record<string, boolean>>,
  groupsAndTypes: Record<string, string[]>,
  search: string
): ImageSummary[] {
  let filtered = summaries.filter((img) => {
    for (const [filter, values] of Object.entries(filters)) {
      const imgValue = img.dynamicFilters[filter];
      if (!imgValue || !values[imgValue]) {
        return false;
      }
    }

    return true;
  });

  filtered = filtered.filter(
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
