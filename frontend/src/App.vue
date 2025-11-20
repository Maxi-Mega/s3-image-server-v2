<script lang="ts" setup>
import Error from "@/components/ErrorBox.vue";
import { fetchStaticInfo } from "@/composables/requests";
import { StaticInfo } from "@/models/static_info";
import { useFilterStore } from "@/stores/filters.ts";
import { useStaticInfoStore } from "@/stores/static_info";
import { plainToInstance } from "class-transformer";
import { type Ref, ref } from "vue";
import { RouterView } from "vue-router";

function setFavicon(base64Data: string): void {
  const link = (document.querySelector("link[rel*='icon']") ||
    document.createElement("link")) as HTMLLinkElement;
  link.type = "image/x-icon";
  link.rel = "icon";
  link.href = "data:image/png;base64," + base64Data;
  document.getElementsByTagName("head")[0]?.appendChild(link);
}

const error: Ref<string | undefined> = ref();

fetchStaticInfo()
  .then((staticInfo) => {
    const info = plainToInstance(StaticInfo, staticInfo);
    useStaticInfoStore().setStaticInfo(info);
    useFilterStore().initFilterModes(info.dynamicFilters);

    if (info.windowTitle) document.title = info.windowTitle;
    if (info.faviconBase64) setFavicon(info.faviconBase64);
  })
  .catch((err) => {
    console.error("Failed to fetch static info:", err);
    if (err instanceof Response) {
      error.value = err.statusText;

      const contentType = err.headers.get("content-type");
      if (contentType && contentType.startsWith("application/json")) {
        err.json().then((data) => {
          console.info("Data:", data);
          if ("error" in data) {
            error.value += " - " + data.error;
          }
        });
      }
    } else {
      error.value = err;
    }
  });
</script>

<template>
  <RouterView v-if="!error" />
  <Error v-else :message="error" standalone />
</template>
