<script setup lang="ts">
import { formatDate, wbr } from "@/composables/images";
import { resolveBackendURL } from "@/composables/url";
import type { ImageSummary } from "@/models/image";
import { useFilterStore } from "@/stores/filters";
import { storeToRefs } from "pinia";

defineProps<{
  summary: ImageSummary;
}>();

const emit = defineEmits<{
  (e: "openModal", img: ImageSummary): void;
}>();

const { globalFontSize } = storeToRefs(useFilterStore());

/*onMounted(() => {
  console.info("Mounted", props.summary.key);
  window.HSStaticMethods.autoInit("overlay");
});

onBeforeUnmount(()=>{
  console.info("Unmounted", props.summary.key);
  window.HSStaticMethods.cleanCollection("overlay");
});*/
</script>

<template>
  <!-- data-hs-overlay="#image-modal" -->
  <div
    class="flex cursor-pointer flex-col justify-end rounded-lg p-2 transition-shadow hover:shadow-lg"
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

<style scoped></style>
