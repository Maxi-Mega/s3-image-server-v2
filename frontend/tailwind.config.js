import { createRequire } from 'module';
const require = createRequire(import.meta.url);

/** @type {import("tailwindcss").Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
    "./node_modules/preline/dist/*.js",
  ],
  theme: {
    extend: {},
  },
  plugins: [
    require("preline/plugin"),
    require("@tailwindcss/forms"),
  ],
};

