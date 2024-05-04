<script setup lang="ts">
import type { ImageSummary } from "@/models/image";
import { resolveBackendURL } from "@/composables/url";
import { onMounted } from "vue";
import { useFilterStore } from "@/stores/filters";
import { storeToRefs } from "pinia";

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

const wbr = (name: string): string => name.replace(/_/g, "<wbr>_");
</script>

<template>
  <div>
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
        <!--      <span
        class="absolute left-0 top-0 rounded-br-lg bg-red-500 px-3 py-1.5 text-sm uppercase tracking-wider text-white"
        >sale</span
      >-->
      </div>
      <p
        v-html="wbr(summary.name)"
        class="mb-1 text-center text-sm text-gray-300 transition duration-100"
        :style="`font-size: ${globalFontSize}`"
      ></p>
    </div>
  </div>
</template>

<style scoped></style>
