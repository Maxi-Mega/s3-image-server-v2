<script lang="ts" setup>
import DropdownIcon from "@/components/icons/DropdownIcon.vue";
import type { ImageGroup } from "@/models/static_info";
import { useFilterStore } from "@/stores/filters";
import { onMounted } from "vue";

defineProps<{
  group: ImageGroup;
}>();

const filterStore = useFilterStore();

onMounted(() => {
  window.HSStaticMethods.autoInit("dropdown");
});

function handleToggle(ev: any, group: string, type?: string): void {
  const typeInputs = Array.from(
    document.querySelectorAll(`input.img-type[type=checkbox][image-group='${group}']`) as NodeList
  ) as Array<HTMLInputElement>;

  if (type) {
    // specific type toggle
    const groupInput = document.querySelector(
      `input.img-group[type=checkbox][image-group='${group}']`
    ) as HTMLInputElement;
    if (typeInputs.every((input) => input.checked)) {
      groupInput.indeterminate = false;
      groupInput.checked = true;
    } else if (typeInputs.every((input) => !input.checked)) {
      groupInput.indeterminate = false;
      groupInput.checked = false;
    } else {
      groupInput.indeterminate = true;
    }
  } else {
    // toggle of the whole group
    const groupChecked = ev.target.checked;
    typeInputs.forEach((el) => (el.checked = groupChecked));
  }

  const checked = typeInputs
    .filter((input) => input.checked)
    .map((input) => input.getAttribute("image-type") as string);
  filterStore.setCheckedTypes(group, checked);
}
</script>

<template>
  <div class="hs-dropdown relative inline-flex flex-nowrap items-center [--auto-close:inside]">
    <input
      :id="`hs-checked-checkbox-${group.name}`"
      type="checkbox"
      class="img-group mt-0.5 shrink-0 rounded border-neutral-700 bg-neutral-800 text-blue-600 checked:border-blue-500 checked:bg-blue-500 focus:ring-blue-500 focus:ring-offset-gray-800 disabled:pointer-events-none disabled:opacity-50"
      :checked="true"
      :image-group="group.name"
      @click="handleToggle($event, group.name)"
    />
    <button
      :title="`bucket '${group.bucket}'`"
      class="hs-dropdown-toggle ml-1 flex w-full cursor-pointer items-center text-lg font-medium text-gray-200 hover:text-gray-100 focus:text-gray-100"
      type="button"
    >
      {{ group.name }}
      <DropdownIcon />
    </button>
    <!-- Group's types -->
    <div
      class="hs-dropdown-menu duration hs-dropdown-open:opacity-100 mt-2 hidden min-w-60 divide-neutral-700 rounded-lg border border-neutral-700 bg-neutral-800 p-2 opacity-0 shadow-md transition-[opacity,margin] before:absolute before:start-0 before:-top-4 before:h-4 before:w-full after:absolute after:start-0 after:-bottom-4 after:h-4 after:w-full"
    >
      <div
        v-for="type in group.types"
        :key="`${group.name}/${type.name}`"
        class="flex items-center gap-x-3.5 rounded-lg px-3 py-2 text-sm text-gray-400 hover:bg-gray-700 hover:text-gray-300 focus:ring-2 focus:ring-blue-500"
      >
        <input
          :id="`hs-checked-checkbox-${group.name}-${type.name}`"
          type="checkbox"
          class="img-type mt-0.5 shrink-0 rounded border-neutral-700 bg-neutral-800 text-blue-600 checked:border-blue-500 checked:bg-blue-500 focus:ring-blue-500 focus:ring-offset-gray-800 disabled:pointer-events-none disabled:opacity-50"
          :checked="true"
          :image-group="group.name"
          :image-type="type.name"
          @click="handleToggle($event, group.name, type.name)"
        />
        <label
          :for="`hs-checked-checkbox-${group.name}-${type.name}`"
          class="ms-3 text-base text-gray-400"
        >
          {{ type.displayName }}
        </label>
      </div>
    </div>
  </div>
</template>

<style scoped></style>
