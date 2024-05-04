import { ref } from "vue";
import { defineStore } from "pinia";
import type { Image, ImageSummary } from "@/models/image";
import { compareSummaries } from "@/composables/images";

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
      this.allSummaries = summaries.sort(compareSummaries).flatMap((e) => [e, e, e, e, e]);
      this.filteredCount = this.allSummaries.length;
    },
    findImage(bucket: string, key: string): Image | undefined {
      return this.allImages.find(
        (img) => img.imageSummary.bucket === bucket && img.imageSummary.key === key
      );
    },
  },
});
