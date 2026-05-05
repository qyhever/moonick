import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";
import { devProxy } from "./devProxy";

export default defineConfig({
  base: "/moonick/",
  plugins: [react()],
  server: {
    proxy: devProxy,
    port: 5080
  },
  test: {
    environment: "jsdom",
    globals: true,
  },
});
