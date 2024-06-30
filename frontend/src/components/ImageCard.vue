<script setup lang="ts">
import type { ImageSummary } from "@/models/image";
import { resolveBackendURL } from "@/composables/url";
import { onMounted } from "vue";
import { useFilterStore } from "@/stores/filters";
import { storeToRefs } from "pinia";
import { formatDate, wbr } from "@/composables/images";

defineProps<{
  summary: ImageSummary;
}>();

const emit = defineEmits<{
  (e: "openModal", img: ImageSummary): void;
}>();

const { globalFontSize } = storeToRefs(useFilterStore());

onMounted(() => {
  window.HSStaticMethods.autoInit("overlay");
});
</script>

<template>
  <div
    class="flex cursor-pointer flex-col justify-end rounded-lg p-2 transition-shadow hover:shadow-lg"
    data-hs-overlay="#image-modal"
    @click="emit('openModal', summary)"
  >
    <div class="group relative mb-2 flex justify-center overflow-hidden rounded-lg lg:mb-3">
      <img
        :src="resolveBackendURL('/api/cache/' + summary.cachedObject.cacheKey)"
        :alt="summary.cachedObject.cacheKey"
        class="h-full w-full object-cover object-center transition duration-300 group-hover:scale-105"
        :title="`${summary.type}\n${summary.key}\n${formatDate(summary._lastModified)}`"
      />
      <div
        v-if="summary.features"
        class="absolute left-0 top-0 flex h-full w-full items-start justify-center overflow-hidden *:hover:bg-transparent *:hover:backdrop-blur-[2px]"
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
      <span v-html="wbr(summary.name)" class="rounded bg-zinc-700/20 p-2 leading-9"></span>
    </p>
  </div>
</template>

<style scoped></style>
