<script lang="ts" setup>
import { computed, ref, watch } from "vue";
import { useQuery } from "@vue/apollo-composable";
import { useStaticInfoStore } from "@/stores/static_info";
import { useImageStore } from "@/stores/images";
import { ALL_IMAGE_SUMMARIES } from "@/models/queries";
import GroupDropdown from "@/components/GroupDropdown.vue";
import LoaderSpinner from "@/components/LoaderSpinner.vue";
import ImageGrid from "@/components/ImageGrid.vue";
import { processSummaries } from "@/composables/images";
import FiltersBar from "@/components/FiltersBar.vue";
import { useFilterStore } from "@/stores/filters";
import type { ImageGroup } from "@/models/static_info";

const staticInfo = useStaticInfoStore();
const imageStore = useImageStore();
const filterStore = useFilterStore();

const groupsAndTypes = computed(() => staticInfo.staticInfo?.imageGroups || ([] as ImageGroup[]));
const { result, loading } = useQuery(ALL_IMAGE_SUMMARIES);

watch(groupsAndTypes, (value: ImageGroup[]) => {
  filterStore.reset();

  // Activate all types by default
  for (const group of value) {
    filterStore.setCheckedTypes(
      group.name,
      group.types.map((type) => type.name)
    );
  }
});

watch(result, (value) => {
  if (value.getAllImageSummaries) {
    imageStore.populateSummaries(processSummaries(value.getAllImageSummaries));
  }
});
</script>

<template>
  <header
    class="fixed z-10 flex w-full flex-wrap bg-slate-500 py-3 text-sm sm:flex-nowrap sm:justify-start"
  >
    <nav aria-label="Global" class="mx-5 w-full px-4 sm:flex sm:items-center sm:justify-between">
      <div class="flex items-center justify-between gap-10">
        <img
          v-if="staticInfo.staticInfo.logoBase64"
          :src="'data:image/png;base64,' + staticInfo.staticInfo.logoBase64"
          alt="App logo"
          class="max-w-32"
        />
        <h1 v-if="staticInfo.staticInfo.applicationTitle" class="text-xl font-bold text-gray-100">
          {{ staticInfo.staticInfo.applicationTitle }}
        </h1>
        <span v-else class="mx-5 px-5"></span>
        <div
          class="flex flex-row items-center gap-5 overflow-x-auto pb-2 sm:mt-0 sm:justify-end sm:overflow-x-visible sm:pb-0 sm:ps-5 [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-gray-300 dark:[&::-webkit-scrollbar-thumb]:bg-slate-500 [&::-webkit-scrollbar-track]:bg-gray-100 dark:[&::-webkit-scrollbar-track]:bg-slate-700 [&::-webkit-scrollbar]:h-2"
        >
          <GroupDropdown v-for="group in groupsAndTypes" :key="group.name" :group="group" />
        </div>
      </div>
      <div
        class="mt-5 flex flex-row items-center gap-5 overflow-x-auto pb-2 sm:mt-0 sm:justify-end sm:overflow-x-visible sm:pb-0 sm:ps-5"
      >
        <FiltersBar />
      </div>
    </nav>
  </header>
  <main class="min-h-screen w-full">
    <Transition name="loading">
      <LoaderSpinner v-if="loading" key="loading-true" :standalone="true">Loading...</LoaderSpinner>
      <ImageGrid v-else key="loading-false" />
    </Transition>
  </main>
</template>

<style scoped>
header {
  background-color: var(--dark-blue);
}

main {
  background-image: linear-gradient(170deg, var(--dark-blue) 50%, var(--aqua-blue));
}

.loading-enter-active,
.loading-leave-active {
  transition: opacity 0.5s;
}

.loading-enter-from,
.loading-leave-to {
  opacity: 0;
}
</style>
