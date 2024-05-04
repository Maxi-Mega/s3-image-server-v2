<script setup lang="ts">
defineProps<{
  standalone: boolean;
}>();

const absoluteClasses = "absolute left-1/2 top-1/2 -translate-y-3/4";
const relativeClasses = "relative text-center";
</script>

<template>
  <div :class="standalone ? absoluteClasses : relativeClasses">
    <div
      class="inline-block size-6 animate-spin rounded-full border-[3px] border-current border-t-transparent text-blue-500"
      role="status"
      aria-label="loading"
    >
      <span class="sr-only"><slot /></span>
    </div>
  </div>
</template>

<style scoped></style>
