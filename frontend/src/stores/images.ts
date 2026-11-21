import type { EventData } from "@/composables/events";
import {
  compareSummaries,
  type GqlImage,
  processImage,
  processSummary,
} from "@/composables/images";
import { type ImageQueryResult, useImageQuery } from "@/composables/queries.ts";
import type { Image, ImageSummary } from "@/models/image";
import { useFilterStore } from "@/stores/filters.ts";
import { useDebounceFn, type UseDebounceFnReturn } from "@vueuse/core";
import { defineStore } from "pinia";
import { type EffectScope, ref } from "vue";

export const useImageStore = defineStore("images", {
  state: () => {
    return {
      allSummaries: ref<ImageSummary[]>([]),
      allImages: ref<Image[]>([]),
      filteredCount: 0,
      bouncingQueries: new Map<string, UseDebounceFnReturn<() => ImageQueryResult>>(),
    };
  },
  getters: {
    totalCount: (state) => state.allSummaries.length,
  },
  actions: {
    populateSummaries(summaries: ImageSummary[]): void {
      this.allSummaries = summaries.sort(compareSummaries);
      this.filteredCount = this.allSummaries.length;

      const filterStore = useFilterStore();
      summaries.forEach((summary) => filterStore.addFilterOptions(summary.dynamicFilters));
    },
    findImage(bucket: string, key: string): Image | undefined {
      return this.allImages.find(
        (img) => img.imageSummary.bucket === bucket && img.imageSummary.key === key
      );
    },
    updateImage(image: Image): Image {
      const summary = image.imageSummary;
      const summaryIdx = findSummaryIndex(summary.bucket, summary.key, this.allSummaries);
      const current = this.allSummaries[summaryIdx];
      if (current && current.lastUpdateOverall > summary.lastUpdateOverall) {
        summary.lastUpdateOverall = current.lastUpdateOverall;
      }

      let idx = this.allImages.findIndex(
        (img) =>
          img.imageSummary.bucket === image.imageSummary.bucket &&
          img.imageSummary.key === image.imageSummary.key
      );
      if (idx < 0) {
        this.allImages.push(image);
        idx = this.allImages.length - 1;
      } else {
        this.allImages[idx] = image;
      }

      if (summaryIdx >= 0) {
        this.allSummaries[summaryIdx] = summary;
      }

      useFilterStore().addFilterOptions(image.imageSummary.dynamicFilters);

      return this.allImages[idx] as Image;
    },
    handleImageDetailsResult(gqlImage: GqlImage | null, debounceKey?: string): void {
      if (!gqlImage || !gqlImage.getImage) {
        return;
      }

      if (debounceKey) {
        this.bouncingQueries.delete(debounceKey);
      }

      this.updateImage(processImage(gqlImage));
    },
    requestImageDetails(bucket: string, key: string, scope: EffectScope | undefined): void {
      const { onResult } = useImageQuery({ bucket: bucket, name: key }, scope);
      onResult((gqlImage: GqlImage | null) => this.handleImageDetailsResult(gqlImage));
    },
    handleEvent(event: EventData, scope: EffectScope | undefined): void {
      if (event.eventType === "ObjectCreated") {
        const { added, updated, needsDebounceUpdate } = handleCreateEvent(event, this.allSummaries);
        if (added) {
          this.allSummaries.push(added);
          useFilterStore().addFilterOptions(added.dynamicFilters);
        } else if (updated) {
          this.allSummaries[
            findSummaryIndex(event.imageBucket, event.imageKey, this.allSummaries)
          ] = updated;
        }

        if (needsDebounceUpdate) {
          const debouncingQuery = this.bouncingQueries.get(
            bouncingQueryKey(event.imageBucket, event.imageKey)
          );
          if (debouncingQuery) {
            debouncingQuery();
          } else {
            const debouncedQuery = useDebounceFn(
              () => useImageQuery({ bucket: event.imageBucket, name: event.imageKey }, scope),
              1000
            );
            debouncedQuery().then((result: ImageQueryResult) => {
              if (result) {
                result.onResult((gqlImage: GqlImage | null) =>
                  this.handleImageDetailsResult(
                    gqlImage,
                    bouncingQueryKey(event.imageBucket, event.imageKey)
                  )
                );
              }
            });
            this.bouncingQueries.set(
              bouncingQueryKey(event.imageBucket, event.imageKey),
              debouncedQuery
            );
          }
        }
      } else if (event.eventType === "ObjectRemoved") {
        const { updated, remove } = handleRemoveEvent(event, this.allSummaries);
        if (remove) {
          const summaryIdx = findSummaryIndex(event.imageBucket, event.imageKey, this.allSummaries);
          if (summaryIdx >= 0) {
            this.allSummaries.splice(summaryIdx, 1);
          }

          const imageIdx = this.allImages.findIndex(
            (img) =>
              img.imageSummary.bucket === event.imageBucket &&
              img.imageSummary.key === event.imageKey
          );
          if (imageIdx >= 0) {
            this.allImages.splice(imageIdx, 1);
          }
        } else if (updated) {
          // TODO
        }
      } else {
        console.warn("Unknown event type", event.eventType);
      }
    },
  },
});

export function findSummaryIndex(bucket: string, key: string, summaries: ImageSummary[]): number {
  return summaries.findIndex((s) => s.bucket === bucket && s.key === key);
}

function handleCreateEvent(
  event: EventData,
  summaries: ImageSummary[]
): {
  added: ImageSummary | null;
  updated: ImageSummary | null;
  needsDebounceUpdate: boolean;
} {
  const summaryIdx = findSummaryIndex(event.imageBucket, event.imageKey, summaries);
  const current = summaries[summaryIdx];
  if (event.objectType === "preview" && event.object) {
    const summary = processSummary(event.object as ImageSummary);
    if (current) {
      if (current.lastUpdateOverall > summary.lastUpdateOverall) {
        summary.lastUpdateOverall = current.lastUpdateOverall;
      }
      return { added: null, updated: summary, needsDebounceUpdate: false };
    }
    return { added: summary, updated: null, needsDebounceUpdate: false };
  }

  if (!current || event.objectType === "preview") {
    return { added: null, updated: null, needsDebounceUpdate: false };
  }

  const updated = {
    ...current,
    _lastModified: event.objectTime,
    _hasBeenUpdated: current._hasBeenUpdated || event.objectType === "target",
    lastUpdateOverall:
      event.objectTime > current.lastUpdateOverall ? event.objectTime : current.lastUpdateOverall,
  };
  const needsDebounceUpdate =
    event.objectType === "dynamic_input" &&
    document.querySelector(`img[full-key='${event.imageBucket}/${event.imageKey}'][src]`) !== null;

  return { added: null, updated, needsDebounceUpdate };
}

function handleRemoveEvent(
  event: EventData,
  summaries: ImageSummary[]
): {
  updated: ImageSummary | null;
  remove: boolean;
} {
  const summaryIdx = findSummaryIndex(event.imageBucket, event.imageKey, summaries);
  if (summaryIdx < 0 || event.objectType === "preview") {
    return { updated: null, remove: true };
  }

  return { updated: summaries[summaryIdx], remove: false };
}

function bouncingQueryKey(bucket: string, key: string): string {
  return `${bucket}\\${key}`;
}
