<script setup lang="ts">
import Error from "@/components/Error.vue";
</script>

<template>
  <h1>404</h1>
  <Error message="404 - Page not found" standalone>
    <h2 class="text-xl">
      <a
        class="text-blue-400 underline decoration-blue-500 underline-offset-2 hover:opacity-90"
        href="/"
        >Go to Home page</a
      >
    </h2>
  </Error>
</template>
