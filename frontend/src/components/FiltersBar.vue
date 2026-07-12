<script setup lang="ts">
import RangeInput from "@/components/RangeInput.vue";
import { useFilterStore } from "@/stores/filters";
import { useImageStore } from "@/stores/images";
import { useStaticInfoStore } from "@/stores/static_info";
import { X } from "@lucide/vue";
import { useTemplateRef } from "vue";

const staticInfo = useStaticInfoStore();
const imageStore = useImageStore();
const filterStore = useFilterStore();

const textSearchInput = useTemplateRef<HTMLInputElement>("filters-text-search");

function clearTextSearchInput() {
  if (textSearchInput.value) {
    (document.getElementById(textSearchInput.value.id) as HTMLInputElement).value = "";
    filterStore.searchQuery = "";
    textSearchInput.value.focus();
  }
}
</script>

<template>
  <div
    class="mt-5 grow-[0.3] items-center gap-1 overflow-x-auto pb-2 sm:mt-0 sm:justify-end sm:overflow-x-visible sm:pb-0"
    style="display: inherit"
  >
    <div class="min-w-12 grow space-y-3 px-1">
      <div class="relative">
        <input
          id="filters-text-search"
          ref="filters-text-search"
          type="search"
          class="block w-full rounded-md border border-gray-100 bg-transparent py-1 pr-5 pl-2 text-base text-gray-200 placeholder-gray-300 placeholder-shown:border-neutral-200 focus:border-blue-500 focus:ring-neutral-600"
          placeholder="Search"
          @input="(e) => (filterStore.searchQuery = (e.target as HTMLInputElement).value)"
        />
        <button
          type="button"
          title="Clear input"
          class="text-muted-foreground focus:text-primary-focus absolute inset-y-0 inset-e-0 z-20 flex cursor-pointer items-center rounded-e-md px-1 focus:outline-hidden"
          @click="clearTextSearchInput"
        >
          <X :size="16" color="var(--color-gray-100)" />
        </button>
      </div>
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
        local-storage-key="main-scaler"
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
