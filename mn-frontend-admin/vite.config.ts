import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";
import { devProxy } from "./devProxy";

export default defineConfig({
  base: "/moonick-admin/",
  plugins: [react()],
  server: {
    proxy: devProxy,
    port: 5090
  },
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: ["./src/test/setup.ts"],
  },
});
