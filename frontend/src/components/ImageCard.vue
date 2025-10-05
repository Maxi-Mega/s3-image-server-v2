<script setup lang="ts">
import { formatDate, wbr } from "@/composables/images";
import { resolveBackendURL } from "@/composables/url";
import type { ImageSummary } from "@/models/image";
import { useFilterStore } from "@/stores/filters";
import { storeToRefs } from "pinia";
import { computed } from "vue";

const props = defineProps<{
  summary: ImageSummary;
  placeholderImageWidth: number;
}>();

const emit = defineEmits<{
  (e: "openModal", img: ImageSummary): void;
}>();

const { globalFontSize } = storeToRefs(useFilterStore());

const imgSize = computed(() => {
  if (!props.summary.size || !props.summary.size.width || !props.summary.size.height) {
    return {
      minWidth: undefined,
      minHeight: undefined,
    };
  }

  return {
    minWidth: Math.round(props.placeholderImageWidth),
    minHeight: Math.round(
      (props.placeholderImageWidth / props.summary.size.width) * props.summary.size.height
    ),
  };
});
</script>

<template>
  <div
    class="flex cursor-pointer flex-col justify-end rounded-lg p-2 transition-shadow hover:shadow-lg"
    @click="emit('openModal', summary)"
  >
    <div class="group relative mb-2 flex justify-center overflow-hidden rounded-lg lg:mb-3">
      <img
        v-lazy-img="{
          src: resolveBackendURL('/api/cache/' + summary.cachedObject.cacheKey),
        }"
        :alt="summary.cachedObject.cacheKey"
        :title="`${summary.type}\n${summary.key}\n${formatDate(summary._lastModified)}`"
        class="skeleton h-full w-full object-cover object-center transition duration-300 group-hover:scale-105"
        :width="imgSize.minWidth"
        :height="imgSize.minHeight"
      />

      <div
        v-if="summary.features"
        class="absolute top-0 left-0 flex h-full w-full items-start justify-center overflow-hidden *:hover:bg-transparent *:hover:backdrop-blur-[2px]"
        :title="`${summary.type}\n${summary.key}\n${formatDate(summary._lastModified)}`"
      >
        <div
          class="mt-2 rounded-md bg-zinc-700/30 p-1 text-center text-green-500 backdrop-blur-sm transition duration-200"
        >
          <span v-if="summary.features.class" class="text-md font-bold">
            {{ summary.features.class }}: {{ summary.features.count }}
          </span>
          <ul class="text-sm">
            <li v-for="(count, feature) in summary.features.objects" :key="feature">
              {{ feature }}: {{ count }}
            </li>
          </ul>
        </div>
      </div>
    </div>
    <p
      class="mb-1 text-center text-sm text-gray-300 transition duration-100"
      :style="`font-size: ${globalFontSize}`"
    >
      <span v-html="wbr(summary.name)" class="rounded p-2"></span>
    </p>
  </div>
</template>

<style scoped>
.skeleton {
  background-color: rgba(121, 134, 161, 0.25);
}
</style>
