<script setup lang="ts">
import { useFilterStore } from "@/stores/filters.ts";
import { ChevronDown } from "lucide-vue-next";
import { computed, onMounted } from "vue";

const props = defineProps<{
  filter: string;
}>();

const filterStore = useFilterStore();
const filterValues = computed(() =>
  Object.keys(filterStore.checkedFilters[props.filter] ?? {}).sort()
);

onMounted(() => {
  window.HSStaticMethods.autoInit("dropdown");
});
</script>

<template>
  <div class="hs-dropdown relative inline-flex flex-nowrap items-center [--auto-close:inside]">
    <button
      class="hs-dropdown-toggle ml-1 flex w-full cursor-pointer items-center text-lg font-medium text-gray-200 hover:text-gray-100 focus:text-gray-100"
      type="button"
    >
      {{ filter }}
      <ChevronDown :size="16" />
    </button>
    <div
      class="hs-dropdown-menu duration hs-dropdown-open:opacity-100 mt-2 hidden max-h-full min-w-60 divide-neutral-700 overflow-auto rounded-lg border border-neutral-700 bg-neutral-800 p-2 opacity-0 shadow-md transition-[opacity,margin] before:absolute before:start-0 before:-top-4 before:h-4 before:w-full"
    >
      <div
        v-for="value in filterValues"
        :key="value"
        class="flex items-center gap-x-3.5 rounded-lg px-3 py-2 text-sm text-gray-400 hover:bg-gray-700 hover:text-gray-300 focus:ring-2 focus:ring-blue-500"
      >
        <input
          :id="`hs-filter-checkbox-${value}`"
          type="checkbox"
          class="img-type mt-0.5 shrink-0 rounded border-neutral-700 bg-neutral-800 text-blue-600 checked:border-blue-500 checked:bg-blue-500 focus:ring-blue-500 focus:ring-offset-gray-800 disabled:pointer-events-none disabled:opacity-50"
          :checked="true"
          @click="filterStore.toggleFilterValue(filter, value)"
        />
        <label :for="`hs-filter-checkbox-${value}`" class="ms-3 text-base text-gray-400">
          {{ value }}
        </label>
      </div>
    </div>
  </div>
</template>

<style scoped></style>
