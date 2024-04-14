<script setup lang="ts">
import { useQuery } from "@vue/apollo-composable";
import gql from "graphql-tag";

const {result} = useQuery(gql`
{
  getAllImageSummaries
}
`)
</script>

<template>
  <main>
    <pre>{{ result }}</pre>
  </main>
</template>
