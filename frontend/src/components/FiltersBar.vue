<script setup lang="ts">
import RangeInput from "@/components/RangeInput.vue";
import { useFilterStore } from "@/stores/filters";
import { useImageStore } from "@/stores/images";
import { useStaticInfoStore } from "@/stores/static_info";

const staticInfo = useStaticInfoStore();
const imageStore = useImageStore();
const filterStore = useFilterStore();
</script>

<template>
  <div
    class="mt-5 items-center gap-1 overflow-x-auto pb-2 sm:mt-0 sm:justify-end sm:overflow-x-visible sm:ps-5 sm:pb-0"
    style="display: inherit"
  >
    <div class="min-w-12 space-y-3 px-1">
      <input
        type="search"
        class="block w-full rounded-md border border-gray-100 bg-transparent px-2 py-1 text-base text-gray-200 placeholder-gray-300 placeholder-shown:border-neutral-200 focus:border-blue-500 focus:ring-neutral-600"
        placeholder="Search"
        @input="(e) => (filterStore.searchQuery = (e.target as HTMLInputElement).value)"
      />
    </div>
    <p
      v-if="staticInfo.staticInfo.maxImagesDisplayCount > imageStore.totalCount"
      class="min-w-32 px-1 text-center text-lg text-white"
    >
      ({{ imageStore.filteredCount }} / {{ imageStore.totalCount }})
    </p>
    <p
      v-else
      class="min-w-32 px-1 text-center text-lg font-bold text-red-500"
      :title="`Max number of images displayed reached (total is ${imageStore.totalCount})`"
    >
      ({{ imageStore.filteredCount }} / {{ staticInfo.staticInfo.maxImagesDisplayCount }})
    </p>
    <div class="flex justify-center pr-2">
      <RangeInput
        v-if="staticInfo.staticInfo.scaleInitialPercentage"
        id="global-scale-range-slider"
        name="Scale images"
        :min="10"
        :max="30"
        :step="1"
        :initial-scale-percentage="staticInfo.staticInfo.scaleInitialPercentage"
        :base-scale="16"
        width-cls="max-w-32"
        @change="filterStore.setGlobalSizes"
      />
    </div>
    <a
      class="min-w-fit pl-2 text-center text-base text-gray-200"
      href="doc"
      target="_blank"
      title="Click to open the documentation"
    >
      {{ staticInfo.staticInfo.softwareVersion }}
    </a>
  </div>
</template>

<style scoped></style>
