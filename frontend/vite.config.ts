import { fileURLToPath, URL } from "node:url";

import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import Markdown from "unplugin-vue-markdown/vite";
import VueDevTools from "vite-plugin-vue-devtools";

// https://vitejs.dev/config/
export default defineConfig({
  base: "/",
  plugins: [
    vue({
      include: [/\.vue$/, /\.md$/],
    }),
    Markdown({
      transforms: {
        before: (src) => {
          // Remove Markdown comments like [//]: # ...
          return src.replace(
            /^\[\/\/]:\s*#.*$/gm,
            ""
          )
        },
      },
    }),
    VueDevTools(),
  ],
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
  server: {
    proxy: {
      "/api": {
        target: "http://localhost:9999",
        ws: true,
      },
    },
  },
});
