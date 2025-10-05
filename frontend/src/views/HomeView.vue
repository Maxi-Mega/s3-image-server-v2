<script lang="ts" setup>
import FiltersBar from "@/components/FiltersBar.vue";
import GroupDropdown from "@/components/GroupDropdown.vue";
import ImageGrid from "@/components/ImageGrid.vue";
import LoaderSpinner from "@/components/LoaderSpinner.vue";
import { parseEventData } from "@/composables/events";
import { processSummaries } from "@/composables/images";
import { ALL_IMAGE_SUMMARIES } from "@/composables/queries";
import { wsURL } from "@/composables/url";
import type { ImageGroup } from "@/models/static_info";
import { useFilterStore } from "@/stores/filters";
import { useImageStore } from "@/stores/images";
import { useStaticInfoStore } from "@/stores/static_info";
import { useQuery } from "@vue/apollo-composable";
import { useWebSocket } from "@vueuse/core";
import { computed, onBeforeUnmount, onMounted, watch } from "vue";

const staticInfo = useStaticInfoStore();
const imageStore = useImageStore();
const filterStore = useFilterStore();

const groupsAndTypes = computed(() => staticInfo.staticInfo?.imageGroups || ([] as ImageGroup[]));
const { result, loading } = useQuery(ALL_IMAGE_SUMMARIES);

const { open, close } = useWebSocket(wsURL, {
  // heartbeat: true,
  autoReconnect: {
    retries: 5,
    delay: 2000,
    onFailed: handleWSConnectionFailure,
  },
  immediate: false,
  // autoClose: false,
  onMessage: handleWSEvent,
  onError: handleWSError,
  onConnected: () => console.info("WS connected"),
});

watch(groupsAndTypes, async (value: ImageGroup[]) => {
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

function handleWSConnectionFailure() {
  console.warn("Can't open WS connection");
  // TODO: emit error ?
}

function handleWSEvent(ws: WebSocket, event: MessageEvent) {
  try {
    event.data
      .trim()
      .split("\n")
      .filter((line: string) => line.trim() !== "")
      .map(parseEventData)
      .forEach(imageStore.handleEvent);
  } catch (e) {
    console.warn(`Error while handling WS event:\nerror=${e}\nevent data=${event.data}`);
    // TODO: emit error ?
  }
}

function handleWSError(ws: WebSocket, event: Event) {
  console.warn("WS error:", event);
}

onMounted(() => {
  open(); // WS connection
});

onBeforeUnmount(() => {
  close(); // WS connection
});

window.onbeforeunload = () => close();
</script>

<template>
  <header
    class="fixed z-10 flex w-full flex-wrap border-b border-gray-300 bg-[var(--dark-blue)] py-3 text-sm sm:flex-nowrap sm:justify-start"
  >
    <nav aria-label="Global" class="mx-5 w-full sm:flex sm:items-center sm:justify-between">
      <div class="flex items-center justify-between gap-10">
        <img
          v-if="staticInfo.staticInfo.logoBase64"
          :src="'data:image/png;base64,' + staticInfo.staticInfo.logoBase64"
          alt="App logo"
          class="max-w-32"
        />
        <h1 v-if="staticInfo.staticInfo.applicationTitle" class="text-2xl font-bold text-gray-100">
          {{ staticInfo.staticInfo.applicationTitle }}
        </h1>
        <span v-else class="mx-5 px-5"></span>
        <div
          class="flex flex-row items-center gap-5 overflow-x-auto pb-2 sm:mt-0 sm:justify-end sm:overflow-x-visible sm:ps-5 sm:pb-0 [&::-webkit-scrollbar]:h-2 [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-gray-200 [&::-webkit-scrollbar-track]:bg-slate-700"
        >
          <GroupDropdown v-for="group in groupsAndTypes" :key="group.name" :group="group" />
        </div>
      </div>
      <div
        class="mt-5 flex flex-row items-center gap-5 overflow-x-auto pb-2 sm:mt-0 sm:justify-end sm:overflow-x-visible sm:ps-5 sm:pb-0"
      >
        <FiltersBar />
      </div>
    </nav>
  </header>
  <main class="min-h-screen w-full bg-fixed">
    <Transition name="loading">
      <LoaderSpinner v-if="loading" key="loading-true" :standalone="true">Loading...</LoaderSpinner>
      <ImageGrid v-else key="loading-false" />
    </Transition>
  </main>
</template>

<style scoped>
main {
  background-image: linear-gradient(170deg, var(--dark-blue) 40%, #7986a1); /* #e5e7eb */
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
