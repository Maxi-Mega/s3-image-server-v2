import "preline/preline";
import "reflect-metadata";
import "./assets/main.css";

import { createPinia } from "pinia";
import type { IStaticMethods } from "preline/preline";
import { createApp } from "vue";

import { apolloClient } from "@/apollo";
import { DefaultApolloClient } from "@vue/apollo-composable";
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
