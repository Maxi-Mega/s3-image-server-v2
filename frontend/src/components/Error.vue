<script setup lang="ts">
defineProps<{
  message: string;
  sub?: string;
}>();
</script>

<template>
  <div class="container m-auto flex h-full w-full justify-center">
    <div
      id="error-content"
      class="m-auto flex h-auto w-auto -translate-y-1/2 flex-col justify-center rounded-xl border border-amber-500 p-10 text-center shadow-sm"
    >
      <h1 class="text-3xl text-amber-400">{{ message }}</h1>
      <hr class="my-5 w-1/2 translate-x-1/2 border-2 border-gray-400" />
      <slot>
        <h2 class="text-xl text-amber-400">
          <template v-if="sub">{{ sub }}</template>
          <template v-else>See the browser/server console for more information</template>
        </h2>
      </slot>
    </div>
  </div>
</template>
