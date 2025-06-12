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
  <div class="min-w-12 space-y-3">
    <input
      type="search"
      class="block w-full rounded-md border border-gray-100 bg-transparent px-2 py-1 text-base text-gray-200 placeholder-gray-300 placeholder-shown:border-neutral-200 focus:border-blue-500 focus:ring-neutral-600"
      placeholder="Search"
      @input="(e) => (filterStore.searchQuery = (e.target as HTMLInputElement).value)"
    />
  </div>
  <p class="min-w-fit text-lg text-white">
    ({{ imageStore.filteredCount }} / {{ imageStore.totalCount }})
  </p>
  <RangeInput
    v-if="staticInfo.staticInfo.scaleInitialPercentage"
    id="global-scale-range-slider"
    name="Scale images"
    :min="10"
    :max="30"
    :step="1"
    :initial-scale-percentage="staticInfo.staticInfo.scaleInitialPercentage"
    :base-scale="16"
    @change="filterStore.setGlobalSizes"
  />
  <p class="min-w-fit text-base text-gray-200">{{ staticInfo.staticInfo.softwareVersion }}</p>
</template>

<style scoped></style>
