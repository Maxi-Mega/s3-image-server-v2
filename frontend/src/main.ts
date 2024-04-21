import "preline/preline";
import "./assets/main.css";

import type { IStaticMethods } from "preline/preline";
import { createApp } from "vue";
import { createPinia } from "pinia";

import { DefaultApolloClient } from "@vue/apollo-composable";
import { apolloClient } from "@/apollo";
import App from "./App.vue";
import router from "./router";

declare global {
  interface Window {
    HSStaticMethods: IStaticMethods;
  }
}

const app = createApp(App);

app.use(createPinia());
app.use(router);
app.provide(DefaultApolloClient, apolloClient);

app.mount("#app");
