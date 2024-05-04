<script setup lang="ts">
import { computed, type Ref, ref, watch } from "vue";
import { storeToRefs } from "pinia";
import { useImageStore } from "@/stores/images";
import { useFilterStore } from "@/stores/filters";
import { summaryKey } from "@/composables/images";
import { applyFilters } from "@/composables/filters";
import type { ImageSummary } from "@/models/image";
import ImageCard from "@/components/ImageCard.vue";
import ImageModal from "@/components/ImageModal.vue";

const imageStore = useImageStore();
const filterStore = useFilterStore();
const filteredSummaries = ref(imageStore.allSummaries);

const { searchQuery, globalScaleValue } = storeToRefs(filterStore);

watch(
  [filterStore.checkedTypes, searchQuery, imageStore.allSummaries],
  ([types, search, allSummaries]) => {
    console.info("Updating filtered summaries", types, search, allSummaries);
    filteredSummaries.value = applyFilters(allSummaries, types, search);
  }
);

const gridColumnsCount = computed(() => Math.round(globalScaleValue.value / 3));

const selectedImage: Ref<ImageSummary | null> = ref(null);

function openModal(img: ImageSummary) {
  selectedImage.value = img;
}
</script>

<template>
  <div key="image-grid" class="container h-full min-h-screen w-full max-w-full p-6">
    <!-- All (filtered) sumamries -->
    <!-- sm:grid-cols-2 lg:grid-cols-4 xl:grid-cols-5 -->
    <div
      class="mt-12 grid gap-x-4 gap-y-8 md:gap-x-6"
      :style="`grid-template-columns: repeat(${gridColumnsCount}, minmax(0, 1fr))`"
    >
      <ImageCard
        v-for="summary in filteredSummaries"
        :key="summaryKey(summary)"
        :summary="summary"
        @openModal="openModal"
      />
    </div>
    <!-- Modal -->
    <ImageModal id="image-modal" :img="selectedImage" />
  </div>
</template>
