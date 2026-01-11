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

function toggleAll(ev: Event) {
  const checked = (ev.target as HTMLInputElement).checked;
  filterStore.setAllFilterValues(props.filter, checked);
  document
    .querySelectorAll(`input[type=checkbox][filter='${props.filter}']`)
    .forEach((el: Element) => ((el as HTMLInputElement).checked = checked));
}

function toggleOne(ev: Event, value: string) {
  filterStore.toggleFilterValue(props.filter, value);
  const mainCheckbox = document.getElementById(
    `hs-checked-checkbox-filter-${props.filter}`
  ) as HTMLInputElement;
  const subCheckboxes = Array.from(
    document.querySelectorAll("input[type=checkbox][filter]")
  ) as Array<HTMLInputElement>;
  if (subCheckboxes.every((input) => input.checked)) {
    mainCheckbox.indeterminate = false;
    mainCheckbox.checked = true;
  } else if (subCheckboxes.every((input) => !input.checked)) {
    mainCheckbox.indeterminate = false;
    mainCheckbox.checked = false;
  } else {
    mainCheckbox.indeterminate = true;
  }
}
</script>

<template>
  <div
    class="hs-dropdown relative inline-flex flex-nowrap items-center gap-x-1 [--auto-close:inside]"
  >
    <input
      :id="`hs-checked-checkbox-filter-${filter}`"
      type="checkbox"
      class="img-group mt-0.5 shrink-0 rounded border-neutral-700 bg-neutral-800 text-blue-600 checked:border-blue-500 checked:bg-blue-500 focus:ring-blue-500 focus:ring-offset-gray-800 disabled:pointer-events-none disabled:opacity-50"
      :checked="true"
      @click="toggleAll($event)"
    />
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
          :filter="filter"
          @click="toggleOne($event, value)"
        />
        <label :for="`hs-filter-checkbox-${value}`" class="ms-3 text-base text-gray-400">
          {{ value }}
        </label>
      </div>
    </div>
  </div>
</template>

<style scoped></style>
