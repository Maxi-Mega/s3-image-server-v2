<script setup lang="ts">
import CloseIcon from "@/components/icons/CloseIcon.vue";
import type { Image, ImageSummary } from "@/models/image";
import { type Ref, ref, toRefs, watch } from "vue";
import Error from "@/components/ErrorBox.vue";
import LoaderSpinner from "@/components/LoaderSpinner.vue";
import { provideApolloClient, useQuery } from "@vue/apollo-composable";
import { getImage } from "@/models/queries";
import { ApolloError, type WatchQueryFetchPolicy } from "@apollo/client";
import { apolloClient } from "@/apollo";
import { base, formatGeonames, processImage, wbr } from "@/composables/images";
import { resolveBackendURL } from "@/composables/url";
import { HSTabs } from "preline/preline";
import RangeInput from "@/components/RangeInput.vue";
import LeftIcon from "@/components/icons/LeftIcon.vue";
import RightIcon from "@/components/icons/RightIcon.vue";
import GeoMap from "@/components/GeoMap.vue";

const props = defineProps<{
  id: string;
  img: ImageSummary | null;
  paginationHints: [boolean, boolean];
}>();

const emit = defineEmits<{
  (e: "navigate", target: "prev" | "next"): void;
}>();

const { img } = toRefs(props);

const image: Ref<Image | null> = ref(null);
const loading = ref(true);
const error: Ref<ApolloError | null> = ref(null);

const showTargetsScaler = ref(true);
const targetsFontSize = ref("13px");
const targetsWidth = ref(12);

watch(img, async (value) => {
  image.value = null;
  loading.value = true;

  const summary = value;
  if (!summary) {
    return;
  }

  const bucket = summary.bucket;
  const key = summary.key;
  // TODO: take eventual image updates into account when defining a fetch policy
  const fetchPolicy: WatchQueryFetchPolicy =
    summary._hasBeenUpdated || error.value ? "network-only" : "cache-first";

  const { onResult, onError } = provideApolloClient(apolloClient)(() =>
    useQuery(getImage(bucket, key), null, { fetchPolicy: fetchPolicy })
  );

  onResult((result) => {
    loading.value = result.loading;
    error.value = result.error || null;

    if (result.data) {
      image.value = processImage(result.data);
      setTimeout(() => {
        window.HSStaticMethods.autoInit("tabs");

        const targetsTab = document.getElementById("modal-targets-item");
        const mapTab = document.getElementById("modal-map-item");
        if (!targetsTab || !mapTab) {
          console.warn("Can't find targets/map tabs");
          return;
        }

        HSTabs.on("change", targetsTab, toggleTargetsMap);
        HSTabs.on("change", mapTab, toggleTargetsMap);
      }, 100);
    }
  });
  onError((err) => {
    console.warn("Image fetch error:", err);
    error.value = err;
  });
});

function navigate(target: "prev" | "next") {
  emit("navigate", target);
}

function toggleTargetsMap(o: { current: string }) {
  const current = o.current.replace("#modal-", "");

  switch (current) {
    case "targets":
      showTargetsScaler.value = true;
      break;
    case "map":
      showTargetsScaler.value = false;
      break;
    default:
      console.warn("Unknown toggle", o.current);
  }
}

function updateTargetsSize(fontSize: string, rawValue: number) {
  targetsFontSize.value = fontSize;
  targetsWidth.value = 17 - Math.round(rawValue / 4);
}

function targetName(target: string): string {
  target = target.slice(target.lastIndexOf("/") + 1);
  target = target.slice(0, target.lastIndexOf("@"));
  return wbr(target);
}
</script>

<template>
  <div
    :id="id"
    class="hs-overlay pointer-events-none fixed start-0 top-0 z-[80] hidden size-full overflow-y-auto overflow-x-hidden"
  >
    <div
      class="m-3 mt-0 h-[calc(100%-3.5rem)] w-[calc(100%-3.5rem)] opacity-0 transition-all ease-out hs-overlay-open:mt-7 hs-overlay-open:opacity-100 hs-overlay-open:duration-500 sm:mx-auto"
    >
      <div
        class="pointer-events-auto flex h-full flex-col overflow-hidden rounded-xl border border-neutral-700 bg-neutral-800 shadow-sm shadow-neutral-700/70"
      >
        <div
          class="flex items-stretch justify-between gap-x-4 border-b border-neutral-700 px-4 py-3"
        >
          <div v-if="image" class="grid w-full grid-cols-4 gap-x-2 text-white">
            <div
              class="brdr-blue col-span-3 grid grid-cols-12 gap-x-4 gap-y-1 rounded-md border-2 p-2 text-sm"
            >
              <span>Name: </span>
              <span class="col-span-11 font-bold">{{ base(image.imageSummary.key) }}</span>
              <span>Type: </span>
              <span class="col-span-2 font-bold">{{ image.imageSummary.type }}</span>
              <span class="col-span-2">Generation date: </span>
              <span class="col-span-5 font-bold">{{ image._lastModified }}</span>
              <span v-if="image.imageSummary.features"
                >{{ image.imageSummary.features.class }}:
              </span>
              <span v-if="image.imageSummary.features" class="font-bold">{{
                image.imageSummary.features.count
              }}</span>
            </div>
            <div
              class="brdr-blue col-span-1 flex grow items-center justify-center rounded-md border-2 p-2"
            >
              {{ formatGeonames(image.geonames) }}
            </div>
          </div>
          <h3 v-else class="font-bold text-white">Loading image info ...</h3>
          <div class="flex flex-col items-center justify-between">
            <button
              type="button"
              class="flex size-7 items-center justify-center rounded-full border border-transparent text-sm font-semibold text-white hover:bg-neutral-700 disabled:pointer-events-none disabled:opacity-50"
              :data-hs-overlay="'#' + id"
            >
              <span class="sr-only">Close</span>
              <CloseIcon />
            </button>
            <nav class="flex flex-row items-center gap-x-1">
              <button
                type="button"
                :disabled="!paginationHints[0]"
                class="inline-flex min-h-[32px] min-w-8 items-center justify-center gap-x-2 rounded-full px-2 py-2 text-sm text-white transition duration-100 hover:bg-gray-100 hover:bg-white/10 focus:bg-gray-100 focus:bg-white/10 focus:outline-none active:bg-white/20 disabled:pointer-events-none disabled:opacity-50"
                @click="navigate('prev')"
              >
                <LeftIcon />
                <span aria-hidden="true" class="sr-only">Previous</span>
              </button>
              <button
                type="button"
                :disabled="!paginationHints[1]"
                class="inline-flex min-h-[32px] min-w-8 items-center justify-center gap-x-2 rounded-full px-2 py-2 text-sm text-white transition duration-100 hover:bg-gray-100 hover:bg-white/10 focus:bg-gray-100 focus:bg-white/10 focus:outline-none active:bg-white/20 disabled:pointer-events-none disabled:opacity-50"
                @click="navigate('next')"
              >
                <span aria-hidden="true" class="sr-only">Next</span>
                <RightIcon />
              </button>
            </nav>
          </div>
        </div>
        <div class="h-full overflow-y-auto p-4">
          <Error v-if="error" :message="error.message" :standalone="false" />
          <LoaderSpinner v-else-if="loading || !image" key="loading-true" :standalone="false">
            Loading image info...
          </LoaderSpinner>
          <div v-else class="grid h-full grid-cols-5 gap-4">
            <a
              :href="resolveBackendURL('/api/cache/' + image.imageSummary.cachedObject.cacheKey)"
              target="_blank"
              class="brdr-blue col-span-2 h-fit w-full rounded-md border-2"
            >
              <img
                :src="resolveBackendURL('/api/cache/' + image.imageSummary.cachedObject.cacheKey)"
                :alt="image.imageSummary.cachedObject.cacheKey"
                class="rounded-md"
              />
            </a>
            <div class="col-span-3 flex flex-col gap-4">
              <div class="row-span-1 grid max-h-full grid-cols-5 gap-4">
                <div
                  class="brdr-blue col-span-4 h-20 overflow-auto rounded-md border-2 text-sm text-white"
                >
                  <ul class="pl-1">
                    <li v-for="[name, link] in image._links" :key="link" class="py-0.5">
                      <a :href="link" target="_blank" class="hover:underline">
                        {{ name }}
                      </a>
                    </li>
                  </ul>
                </div>
                <div class="col-span-1 flex flex-col items-center justify-between gap-2">
                  <div class="flex h-fit rounded-lg bg-neutral-700 p-1 transition hover:opacity-90">
                    <nav class="flex justify-center space-x-1" aria-label="Tabs" role="tablist">
                      <button
                        type="button"
                        class="active inline-flex items-center gap-x-2 rounded-lg bg-transparent px-4 py-3 text-center text-sm font-medium text-neutral-200 transition disabled:pointer-events-none disabled:opacity-50 hs-tab-active:bg-neutral-800 hs-tab-active:text-white"
                        id="modal-targets-item"
                        data-hs-tab="#modal-targets"
                        aria-controls="modal-targets"
                        role="tab"
                      >
                        Targets
                      </button>
                      <button
                        type="button"
                        :disabled="!image.localization"
                        class="inline-flex items-center gap-x-2 rounded-lg bg-transparent px-4 py-3 text-center text-sm font-medium text-neutral-200 transition disabled:pointer-events-none disabled:opacity-50 hs-tab-active:bg-neutral-800 hs-tab-active:text-white"
                        id="modal-map-item"
                        data-hs-tab="#modal-map"
                        aria-controls="modal-map"
                        role="tab"
                      >
                        Map
                      </button>
                    </nav>
                  </div>
                  <div class="m-auto max-w-[calc(75%)]">
                    <div v-show="showTargetsScaler">
                      <RangeInput
                        id="targets-scale-range-slider"
                        name="Scale targets"
                        :min="5"
                        :max="30"
                        :step="1"
                        :initial-scale-percentage="50"
                        :base-scale="16"
                        :full-width="true"
                        @change="updateTargetsSize"
                      />
                    </div>
                  </div>
                </div>
              </div>
              <div class="brdr-blue row-span-3 h-full rounded-md border-2">
                <div class="h-full p-1">
                  <div id="modal-targets" role="tabpanel" aria-labelledby="modal-targets-item">
                    <div class="flex flex-row flex-wrap justify-around gap-2">
                      <div
                        v-for="target in image.targetFiles"
                        :key="target"
                        :style="`width: ${targetsWidth}rem`"
                      >
                        <a
                          :href="resolveBackendURL('/api/cache/' + target)"
                          target="_blank"
                          class="rounded *:hover:underline"
                        >
                          <img
                            :src="resolveBackendURL('/api/cache/' + target)"
                            :alt="target"
                            class="h-auto rounded object-cover"
                          />
                          <p
                            v-html="targetName(target)"
                            class="mb-1 break-all text-center text-gray-300 transition duration-100"
                            :style="`font-size: ${targetsFontSize}`"
                          ></p>
                        </a>
                      </div>
                      <p v-if="image.targetFiles.length === 0" class="text-sm text-white">
                        No targets available
                      </p>
                    </div>
                  </div>
                  <div
                    id="modal-map"
                    class="hidden h-full"
                    role="tabpanel"
                    aria-labelledby="modal-map-item"
                  >
                    <GeoMap v-if="image.localization" :localization="image.localization" />
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.brdr-blue {
  border-color: var(--aqua-blue);
}
</style>
