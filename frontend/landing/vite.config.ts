import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

export default defineConfig({
  plugins: [react(), tailwindcss()],
  // Relative base so the same build works at both:
  //   - the custom domain root (https://chronicleclassic.com/)
  //   - the github.io subpath (https://<user>.github.io/chronicle/)
  base: "./",
  build: {
    outDir: "dist",
  },
});
