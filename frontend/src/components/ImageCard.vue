<script setup lang="ts">
import type { ImageSummary } from "@/models/image";
import { resolveBackendURL } from "@/composables/url";
import { onMounted } from "vue";
import { useFilterStore } from "@/stores/filters";
import { storeToRefs } from "pinia";
import { wbr } from "@/composables/images";

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
    class="cursor-pointer rounded-lg p-2 transition-shadow hover:shadow-lg"
    data-hs-overlay="#image-modal"
    @click="emit('openModal', summary)"
  >
    <div class="group relative mb-2 flex justify-center overflow-hidden rounded-lg lg:mb-3">
      <img
        :src="resolveBackendURL('/api/cache/' + summary.cacheKey)"
        :alt="summary.cacheKey"
        class="h-full w-full object-cover object-center transition duration-300 group-hover:scale-105"
      />
      <div
        v-if="summary.features"
        class="absolute left-0 top-4 flex h-full w-full items-start justify-center overflow-hidden *:hover:bg-opacity-75"
      >
        <div class="rounded bg-zinc-700 p-1 text-center text-green-500 transition duration-200">
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
      v-html="wbr(summary.name)"
      class="mb-1 text-center text-sm text-gray-300 transition duration-100"
      :style="`font-size: ${globalFontSize}`"
    ></p>
  </div>
</template>

<style scoped></style>
