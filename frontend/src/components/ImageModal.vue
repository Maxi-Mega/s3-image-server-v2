<script setup lang="ts">
import CloseIcon from "@/components/icons/CloseIcon.vue";
import type { ImageSummary } from "@/models/image";
import { computed, type Ref, ref, watch } from "vue";
import { useImageStore } from "@/stores/images";
import Error from "@/components/Error.vue";
import LoaderSpinner from "@/components/LoaderSpinner.vue";
import { provideApolloClient, useQuery } from "@vue/apollo-composable";
import { getImage } from "@/models/queries";
import { ApolloError } from "@apollo/client";
import { apolloClient } from "@/apollo";
import { processImage } from "@/composables/images";

const props = defineProps<{
  id: string;
  img: ImageSummary | null;
}>();

const imageStore = useImageStore();

let result: Ref<any | undefined> = ref();
let loading = ref(true);
let error: Ref<ApolloError | null>;

const image = computed(() => {
  if (result.value) {
    return processImage(result.value);
  }

  return null;
});

watch(props, async (value) => {
  result.value = undefined;
  loading.value = true;

  if (!value.img) {
    return;
  }

  let img = imageStore.findImage(value.img.bucket, value.img.key);
  if (!img) {
    const bucket = value.img.bucket;
    const key = value.img.key;
    ({ result, loading, error } = provideApolloClient(apolloClient)(() =>
      useQuery(getImage(bucket, key))
    ));
  }
});
</script>

<template>
  <div
    :id="id"
    class="hs-overlay pointer-events-none fixed start-0 top-0 z-[80] hidden size-full overflow-y-auto overflow-x-hidden"
  >
    <div
      class="m-3 mt-0 opacity-0 transition-all ease-out hs-overlay-open:mt-7 hs-overlay-open:opacity-100 hs-overlay-open:duration-500 lg:mx-auto lg:w-full lg:max-w-4xl"
    >
      <div
        class="pointer-events-auto flex flex-col rounded-xl border bg-white shadow-sm dark:border-neutral-700 dark:bg-neutral-800 dark:shadow-neutral-700/70"
      >
        <div class="flex items-center justify-between border-b px-4 py-3 dark:border-neutral-700">
          <h3 class="font-bold text-gray-800 dark:text-white">
            {{ img ? img.name : "You should reload the page ..." }}
          </h3>
          <button
            type="button"
            class="flex size-7 items-center justify-center rounded-full border border-transparent text-sm font-semibold text-gray-800 hover:bg-gray-100 disabled:pointer-events-none disabled:opacity-50 dark:text-white dark:hover:bg-neutral-700"
            :data-hs-overlay="'#' + id"
          >
            <span class="sr-only">Close</span>
            <CloseIcon />
          </button>
        </div>
        <div class="overflow-y-auto p-4">
          <Error v-if="error" :message="error.message" :standalone="false" />
          <LoaderSpinner v-else-if="loading || !image" key="loading-true" :standalone="false"
            >Loading info...</LoaderSpinner
          >
          <pre v-else class="mt-1 text-gray-800 dark:text-neutral-400"
            >{{ image }}
          </pre>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped></style>
