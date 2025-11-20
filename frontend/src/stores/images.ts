import type { EventData } from "@/composables/events";
import { compareSummaries, type GqlImage, processImage } from "@/composables/images";
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
      bouncingQueries: new Map<[string, string], UseDebounceFnReturn<() => ImageQueryResult>>(),
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

      const summaryIdx = findSummaryIndex(
        image.imageSummary.bucket,
        image.imageSummary.key,
        this.allSummaries
      );
      if (summaryIdx >= 0) {
        this.allSummaries[summaryIdx] = image.imageSummary;
      }

      useFilterStore().addFilterOptions(image.imageSummary.dynamicFilters);

      return this.allImages[idx] as Image;
    },
    handleImageDetailsResult(gqlImage: GqlImage | null, debounceID?: [string, string]): void {
      if (!gqlImage || !gqlImage.getImage) {
        return;
      }

      if (debounceID) {
        this.bouncingQueries.delete(debounceID);
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
          const debouncingQuery = this.bouncingQueries.get([event.imageBucket, event.imageKey]);
          if (debouncingQuery) {
            debouncingQuery();
          } else {
            const debouncedQuery = useDebounceFn(
              () => useImageQuery({ bucket: event.imageBucket, name: event.imageKey }, scope),
              1000
            );
            debouncedQuery().then((result: ImageQueryResult) => {
              result.onResult((gqlImage: GqlImage | null) =>
                this.handleImageDetailsResult(gqlImage, [event.imageBucket, event.imageKey])
              );
            });
            this.bouncingQueries.set([event.imageBucket, event.imageKey], debouncedQuery);
          }
        }
      } else if (event.eventType === "ObjectRemoved") {
        const { updated, remove } = handleRemoveEvent(event, this.allSummaries);
        if (remove) {
          this.allSummaries = this.allSummaries.filter(
            (s) => s.bucket != event.imageBucket || s.key != event.imageKey
          );
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
  let added: ImageSummary | null = null,
    updated: ImageSummary | null = null,
    needsDebounceUpdate: boolean = false;

  const summaryIdx = findSummaryIndex(event.imageBucket, event.imageKey, summaries);
  if (summaryIdx < 0) {
    if (event.objectType !== "preview") {
      return { added: null, updated: null, needsDebounceUpdate: false };
    }

    added = {
      // @ts-expect-error no worries
      ...(event.object as ImageSummary),
      _hasBeenUpdated: false,
      _lastModified: event.objectTime,
    };
  } else {
    // @ts-expect-error no worries
    updated = summaries[summaryIdx];
    switch (event.objectType) {
      case "preview":
        // @ts-expect-error no worries
        updated = event.object;
        break;
      case "dynamic_input":
        if (document.querySelector(`img[full-key='${event.imageBucket}/${event.imageKey}'][src]`)) {
          needsDebounceUpdate = true;
        }
        break;
      default:
        // @ts-expect-error no worries
        updated._hasBeenUpdated = true; // will be fetched on next modal open
    }
    // @ts-expect-error updated is not null
    updated._lastModified = event.objectTime;
  }

  return { added, updated, needsDebounceUpdate };
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

  // @ts-expect-error no worries
  return { updated: summaries[summaryIdx], remove: false };
}
