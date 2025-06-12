<script setup lang="ts">
import ImageCard from "@/components/ImageCard.vue";
import ImageModal from "@/components/ImageModal.vue";
import { applyFilters } from "@/composables/filters";
import { compareSummaries, limitDisplayedImages, summaryKey } from "@/composables/images";
import type { ImageSummary } from "@/models/image";
import { useFilterStore } from "@/stores/filters";
import { findSummaryIndex, useImageStore } from "@/stores/images";
import { useStaticInfoStore } from "@/stores/static_info";
import { storeToRefs } from "pinia";
import { HSOverlay } from "preline";
import { computed, nextTick, onMounted, type Ref, ref, watch } from "vue";

const staticInfoStore = useStaticInfoStore();
const imageStore = useImageStore();
const filterStore = useFilterStore();

const filteredSummaries = ref<ImageSummary[]>([]);
const { searchQuery, globalScaleValue } = storeToRefs(filterStore);

const modalOverlay = ref<HSOverlay | undefined>();

onMounted(() => {
  const modalEl = document.getElementById("image-modal");
  if (modalEl) {
    modalOverlay.value = new HSOverlay(modalEl);
  } else {
    console.warn("Modal element not found");
  }

  filteredSummaries.value = limitDisplayedImages(
    imageStore.allSummaries.sort(compareSummaries),
    staticInfoStore.staticInfo
  );
  imageStore.filteredCount = filteredSummaries.value.length;
});

watch(
  [filterStore.checkedTypes, searchQuery, imageStore.allSummaries],
  ([types, search, allSummaries]) => {
    filteredSummaries.value = limitDisplayedImages(
      applyFilters(allSummaries, types, search),
      staticInfoStore.staticInfo
    );
    imageStore.filteredCount = filteredSummaries.value.length;
  }
);

const gridColumnsCount = computed(() => Math.round(globalScaleValue.value / 3));

const selectedImage: Ref<ImageSummary | undefined> = ref();
const modalPaginationHints: Ref<[boolean, boolean]> = ref([false, false]);

function openModal(img: ImageSummary, reset = true) {
  if (reset) {
    selectedImage.value = undefined;
  }

  const imgIndex = findSummaryIndex(img.bucket, img.key, filteredSummaries.value);
  if (imgIndex < 0) {
    console.warn("Can't find image to modalize.");
    return;
  }

  if (modalOverlay.value) {
    if (modalOverlay.value.el.classList.contains("hidden")) {
      modalOverlay.value.open();
    }
  }

  // Let the time for the modal to realize that the selected image is possibly undefined
  nextTick(() => {
    selectedImage.value = img;
    modalPaginationHints.value = [imgIndex > 0, imgIndex < filteredSummaries.value.length - 1];
  });
}

function modalNavigate(target: "prev" | "next") {
  const img = selectedImage.value;
  if (!img) {
    return;
  }

  const currentIndex = findSummaryIndex(img.bucket, img.key, filteredSummaries.value);
  if (currentIndex < 0) {
    console.warn("Can't find current modalized image.");
    // selectedImage.value = undefined;
    return;
  }

  const navigatedIndex = currentIndex + (target === "prev" ? -1 : 1);
  if (navigatedIndex < 0 || navigatedIndex >= filteredSummaries.value.length) {
    return;
  }

  openModal(filteredSummaries.value[navigatedIndex], false);
}
</script>

<template>
  <div key="image-grid" class="container h-full min-h-screen w-full max-w-full p-6">
    <!-- Modal -->
    <ImageModal
      id="image-modal"
      :img="selectedImage"
      :pagination-hints="modalPaginationHints"
      @navigate="modalNavigate"
    />
    <!-- All (filtered) sumaries -->
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
  </div>
</template>
