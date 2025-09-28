import type { EventData } from "@/composables/events";
import { compareSummaries } from "@/composables/images";
import type { Image, ImageSummary } from "@/models/image";
import { defineStore } from "pinia";
import { ref } from "vue";

export const useImageStore = defineStore("images", {
  state: () => {
    return {
      allSummaries: ref<ImageSummary[]>([]),
      allImages: ref<Image[]>([]),
      filteredCount: 0,
    };
  },
  getters: {
    totalCount: (state) => state.allSummaries.length,
  },
  actions: {
    populateSummaries(summaries: ImageSummary[]): void {
      this.allSummaries = summaries.sort(compareSummaries);
      this.filteredCount = this.allSummaries.length;
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

      return this.allImages[idx] as Image;
    },
    handleEvent(event: EventData): void {
      if (event.eventType === "ObjectCreated") {
        const { added, updated } = handleCreateEvent(event, this.allSummaries);
        if (added) {
          this.allSummaries.push(added);
        } else if (updated) {
          this.allSummaries[
            findSummaryIndex(event.imageBucket, event.imageKey, this.allSummaries)
          ] = updated;
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
} {
  const summaryIdx = findSummaryIndex(event.imageBucket, event.imageKey, summaries);
  let added: ImageSummary | null = null,
    updated: ImageSummary | null = null;
  if (summaryIdx < 0) {
    if (event.objectType !== "preview") {
      return { added: null, updated: null };
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
      case "geonames":
        // @ts-expect-error no worries
        if (event.object.topLevel) {
          // @ts-expect-error updated is not null
          updated.name = event.object.topLevel;
        } else {
          // @ts-expect-error no worries
          updated.name = event.imageKey;

          // @ts-expect-error no worries
          const lastSlash = updated.name.lastIndexOf("/");
          if (lastSlash > -1) {
            // @ts-expect-error no worries
            updated.name = updated.name.substring(lastSlash + 1);
          }
        }

        // @ts-expect-error no worries
        updated._hasBeenUpdated = true; // the whole geonames object still needs to be fetched
        break;
      case "features":
        // @ts-expect-error updated is not null
        updated.features = event.object;
        break;
      default:
        // @ts-expect-error no worries
        updated._hasBeenUpdated = true; // will be fetched on next modal open
    }
    // @ts-expect-error updated is not null
    updated._lastModified = event.objectTime;
  }

  return { added, updated };
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
