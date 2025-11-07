<script setup lang="ts">
import type { DynamicData } from "@/models/dynamic_data.ts";
import { HSLayoutSplitter } from "preline";
import { onMounted } from "vue";

defineProps<{
  data: DynamicData;
}>();

onMounted(() => HSLayoutSplitter.autoInit());
</script>

<template>
  <div
    class="h-full w-full border-l border-l-neutral-700"
    data-hs-layout-splitter='{
  "verticalSplitterTemplate": "<div><span class=\"absolute top-1/2 start-1/2 -translate-x-1/2 -translate-y-1/2 block w-6 h-4 flex justify-center items-center bg-white text-gray-400 rounded-md cursor-row-resize hover:bg-gray-100\"><svg class=\"shrink-0 size-3.5\" xmlns=\"http://www.w3.org/2000/svg\" width=\"24\" height=\"24\" viewBox=\"0 0 24 24\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\"><circle cx=\"12\" cy=\"9\" r=\"1\"/><circle cx=\"19\" cy=\"9\" r=\"1\"/><circle cx=\"5\" cy=\"9\" r=\"1\"/><circle cx=\"12\" cy=\"15\" r=\"1\"/><circle cx=\"19\" cy=\"15\" r=\"1\"/><circle cx=\"5\" cy=\"15\" r=\"1\"/></svg></span></div>",
  "verticalSplitterClasses": "relative flex border-t border-neutral-700"
}'
  >
    <div class="flex h-full flex-col" data-hs-layout-splitter-vertical-group="">
      <div class="overflow-hidden pl-1" data-hs-layout-splitter-item="50.0" style="flex: 50 1 0">
        <h2 class="text-lg underline">File selectors</h2>
        <div class="flex h-[85%] flex-col gap-1 overflow-auto p-1 text-gray-800">
          <pre
            v-for="[name, sel] in Object.entries(data.fileSelectors)"
            :key="name"
            class="dynamic-item"
            >{{ name }} : {{ sel }}</pre
          >
        </div>
      </div>
      <div class="overflow-hidden pl-1" data-hs-layout-splitter-item="50.0" style="flex: 50 1 0">
        <h2 class="text-lg underline">Expressions</h2>
        <div class="flex h-[85%] flex-col gap-1 overflow-auto p-1 text-gray-800">
          <pre
            v-for="[name, expr] in Object.entries(data.expressions)"
            :key="name"
            class="dynamic-item"
            >{{ name }} : {{ expr.replace(/\\n/gi, "\n") }}</pre
          >
        </div>
      </div>
    </div>
  </div>
</template>
