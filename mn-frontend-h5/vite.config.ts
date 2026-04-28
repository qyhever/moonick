import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";
import { devProxy } from "./devProxy";

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: devProxy,
  },
  test: {
    environment: "jsdom",
    globals: true,
  },
});
