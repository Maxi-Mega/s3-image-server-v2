<script lang="ts" setup>
import { RouterView } from "vue-router";
import Error from "@/components/ErrorBox.vue";
import { fetchStaticInfo } from "@/composables/requests";
import { useStaticInfoStore } from "@/stores/static_info";
import { type Ref, ref } from "vue";

function setFavicon(base64Data: string): void {
  const link = (document.querySelector("link[rel*='icon']") ||
    document.createElement("link")) as HTMLLinkElement;
  link.type = "image/x-icon";
  link.rel = "icon";
  link.href = "data:image/png;base64," + base64Data;
  document.getElementsByTagName("head")[0].appendChild(link);
}

const error: Ref<string | undefined> = ref();

fetchStaticInfo()
  .then((staticInfo) => {
    useStaticInfoStore().setStaticInfo(staticInfo);

    if (staticInfo.windowTitle) document.title = staticInfo.windowTitle;
    if (staticInfo.faviconBase64) setFavicon(staticInfo.faviconBase64);
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
