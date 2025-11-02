<script setup lang="ts">
import Error from "@/components/ErrorBox.vue";
import GeoMap from "@/components/GeoMap.vue";
import LoaderSpinner from "@/components/LoaderSpinner.vue";
import RangeInput from "@/components/RangeInput.vue";
import CloseIcon from "@/components/icons/CloseIcon.vue";
import LeftIcon from "@/components/icons/LeftIcon.vue";
import RightIcon from "@/components/icons/RightIcon.vue";
import { base, formatGeonames, processImage, wbr } from "@/composables/images";
import { useImageQuery } from "@/composables/queries";
import { resolveBackendURL } from "@/composables/url";
import type { Image, ImageSummary } from "@/models/image";
import type { Localization } from "@/models/localization";
import { useImageStore } from "@/stores/images";
import { HSTabs } from "preline/preline";
import { nextTick, reactive, type Ref, ref, toRefs, watch } from "vue";

const props = defineProps<{
  id: string;
  img: ImageSummary | undefined;
  paginationHints: [boolean, boolean];
}>();

const emit = defineEmits<{
  (e: "navigate", target: "prev" | "next"): void;
}>();

const imageStore = useImageStore();

const { img } = toRefs(props);

const image: Ref<Image | null> = ref(null);
const hasTargets = ref(false);
const hasMap = ref(false);

const showingTargets = ref(true);
const targetsFontSize = ref("13px");
const targetsWidth = ref(12);

const enableQuery = ref(false);
const imageQueryVariables = reactive({
  bucket: "",
  name: "",
});

watch(img, async (value) => {
  image.value = null;
  imageQueryVariables.bucket = "";
  imageQueryVariables.name = "";
  enableQuery.value = false;
  loading.value = true;

  const summary = value;
  if (!summary) {
    return;
  }

  const bucket = summary.bucket;
  const key = summary.key;
  const imageFromCache = imageStore.findImage(bucket, key);
  if (imageFromCache && !summary._hasBeenUpdated) {
    console.log("No fetch required");
    image.value = imageFromCache;
    hasTargets.value = image.value.targetFiles.length > 0;
    hasMap.value = image.value.localization != null;
    loading.value = false;
    updateTabs();
    return;
  }

  imageQueryVariables.bucket = bucket;
  imageQueryVariables.name = key;
  enableQuery.value = true;
});

const { loading, error, data } = useImageQuery(imageQueryVariables);

watch(data, (gqlImage) => {
  if (!gqlImage || !gqlImage.getImage) {
    return;
  }

  image.value = imageStore.updateImage(processImage(gqlImage));
  hasTargets.value = image.value.targetFiles.length > 0;
  hasMap.value = image.value.localization != null;
  updateTabs();
});

function updateTabs() {
  if (hasTargets.value) {
    showingTargets.value = true;
  }

  if (hasTargets.value || hasMap.value) {
    nextTick(() => {
      window.HSStaticMethods.autoInit("tabs");

      const targetsTab = document.getElementById("modal-targets-item");
      const mapTab = document.getElementById("modal-map-item");
      if (!targetsTab || !mapTab) {
        console.warn("Can't find targets/map tabs");
        return;
      }

      HSTabs.on("change", targetsTab, toggleTargetsMap);
      HSTabs.on("change", mapTab, toggleTargetsMap);

      if (!hasTargets.value && hasMap.value) {
        HSTabs.open(mapTab);
      }
    });
  }
}

function navigate(target: "prev" | "next") {
  emit("navigate", target);
}

function toggleTargetsMap(o: { current: string }) {
  const current = o.current.replace("#modal-", "");

  switch (current) {
    case "targets":
      showingTargets.value = true;
      break;
    case "map":
      showingTargets.value = false;
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
  <span :id="`${id}-trigger`" class="hidden" data-hs-overlay="#image-modal"></span>
  <div
    :id="id"
    class="hs-overlay pointer-events-none fixed start-0 top-0 z-[80] hidden size-full overflow-x-hidden overflow-y-auto"
  >
    <div
      class="hs-overlay-open:mt-7 hs-overlay-open:opacity-100 hs-overlay-open:duration-500 m-3 mt-0 h-[calc(100%-3.5rem)] w-[calc(100%-3.5rem)] opacity-0 transition-all ease-out sm:mx-auto"
    >
      <div
        class="pointer-events-auto flex h-full flex-col overflow-hidden rounded-xl border border-neutral-700 bg-gray-200 shadow-sm shadow-neutral-700/70"
      >
        <div
          class="flex items-stretch justify-between gap-x-4 border-b border-neutral-700 px-4 py-3"
        >
          <div v-if="image" class="grid w-full grid-cols-4 gap-x-2 text-white">
            <div
              class="bg-blue col-span-3 grid grid-cols-12 gap-x-4 gap-y-1 rounded-md p-2 text-sm"
            >
              <span>Name: </span>
              <span class="col-span-11 font-bold">{{ base(image.imageSummary.key) }}</span>
              <span>Type: </span>
              <span class="col-span-3 font-bold">{{ image.imageSummary.type }}</span>
              <span class="col-span-2 text-end">Generation date: </span>
              <span class="col-span-4 font-bold">{{ image._lastModified }}</span>
              <span v-if="image.imageSummary.productInfo?.summary" class="col-span-2 text-center">{{
                image.imageSummary.productInfo.summary
              }}</span>
            </div>
            <div class="bg-blue col-span-1 flex grow items-center justify-center rounded-md p-2">
              {{ formatGeonames(image.imageSummary.geonames) }}
            </div>
          </div>
          <h3 v-else class="font-bold text-white">Loading image info ...</h3>
          <div class="flex flex-col items-center justify-between">
            <button
              type="button"
              class="flex size-7 cursor-pointer items-center justify-center rounded-full border border-transparent text-sm font-semibold text-gray-700 hover:bg-gray-400 disabled:pointer-events-none disabled:opacity-50"
              :data-hs-overlay="'#' + id"
            >
              <span class="sr-only">Close</span>
              <CloseIcon />
            </button>
            <nav class="flex flex-row items-center gap-x-1">
              <button
                type="button"
                :disabled="!paginationHints[0]"
                class="inline-flex min-h-[32px] min-w-8 cursor-pointer items-center justify-center gap-x-2 rounded-full px-2 py-2 text-sm text-gray-700 transition duration-100 hover:bg-gray-300 focus:bg-gray-100 focus:bg-white/10 focus:outline-none active:bg-white/20 disabled:pointer-events-none disabled:opacity-50"
                @click="navigate('prev')"
              >
                <LeftIcon />
                <span aria-hidden="true" class="sr-only">Previous</span>
              </button>
              <button
                type="button"
                :disabled="!paginationHints[1]"
                class="inline-flex min-h-[32px] min-w-8 cursor-pointer items-center justify-center gap-x-2 rounded-full px-2 py-2 text-sm text-gray-700 transition duration-100 hover:bg-gray-300 focus:bg-gray-100 focus:bg-white/10 focus:outline-none active:bg-white/20 disabled:pointer-events-none disabled:opacity-50"
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
              class="col-span-2 flex h-fit max-h-full w-full justify-center rounded-md"
            >
              <img
                :src="resolveBackendURL('/api/cache/' + image.imageSummary.cachedObject.cacheKey)"
                :alt="image.imageSummary.cachedObject.cacheKey"
                class="max-h-[80svh] rounded-md"
              />
            </a>
            <div :key="image.imageSummary.key" class="col-span-3 flex flex-col gap-4">
              <div class="row-span-1 grid max-h-full grid-cols-5 gap-4">
                <div
                  v-if="image._links.length > 0"
                  class="bg-blue col-span-4 h-20 overflow-auto rounded-md text-sm text-white"
                >
                  <ul class="pl-1">
                    <li v-for="[name, link] in image._links" :key="link" class="py-0.5">
                      <a :href="link" target="_blank" class="hover:underline">
                        {{ name }}
                      </a>
                    </li>
                  </ul>
                </div>
                <div
                  v-show="hasTargets || hasMap"
                  class="col-span-1 flex flex-col items-center justify-between gap-2"
                >
                  <div class="flex h-fit rounded-lg bg-gray-300 p-1 transition hover:opacity-90">
                    <nav class="flex justify-center space-x-1" aria-label="Tabs" role="tablist">
                      <button
                        type="button"
                        :disabled="!hasTargets"
                        class="active hs-tab-active:bg-[var(--dark-blue)] hs-tab-active:text-gray-200 inline-flex cursor-pointer items-center gap-x-2 rounded-lg bg-transparent px-4 py-3 text-center text-sm font-medium text-gray-700 transition disabled:pointer-events-none disabled:opacity-50"
                        id="modal-targets-item"
                        data-hs-tab="#modal-targets"
                        aria-controls="modal-targets"
                        role="tab"
                      >
                        Targets
                      </button>
                      <button
                        type="button"
                        :disabled="!hasMap"
                        class="hs-tab-active:bg-[var(--dark-blue)] hs-tab-active:text-gray-200 inline-flex cursor-pointer items-center gap-x-2 rounded-lg bg-transparent px-4 py-3 text-center text-sm font-medium text-gray-700 transition disabled:pointer-events-none disabled:opacity-50"
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
                    <div
                      v-show="showingTargets"
                      class="flex flex-col items-center justify-center gap-y-2"
                    >
                      <span class="text-sm text-gray-700"
                        >Count: {{ image.targetFiles.length }}</span
                      >
                      <RangeInput
                        id="targets-scale-range-slider"
                        name="Scale targets"
                        :min="5"
                        :max="30"
                        :step="1"
                        :initial-scale-percentage="50"
                        :base-scale="16"
                        widthCls="w-full"
                        @change="updateTargetsSize"
                      />
                    </div>
                  </div>
                </div>
              </div>
              <div v-show="hasTargets || hasMap" class="row-span-3 mb-3 h-full rounded-md">
                <div
                  id="modal-targets"
                  role="tabpanel"
                  aria-labelledby="modal-targets-item"
                  class="active bg-blue"
                >
                  <div class="flex flex-row flex-wrap justify-around gap-2 py-2">
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
                          class="mb-1 text-center break-all text-gray-300 transition duration-100"
                          :style="`font-size: ${targetsFontSize}`"
                        ></p>
                      </a>
                    </div>
                    <p v-if="!hasTargets" class="text-sm text-white">No targets available</p>
                  </div>
                </div>
                <div
                  id="modal-map"
                  class="hidden h-full"
                  role="tabpanel"
                  aria-labelledby="modal-map-item"
                >
                  <GeoMap v-if="hasMap" :localization="image.localization as Localization" />
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
.bg-blue {
  background-color: var(--dark-blue);
}
</style>
