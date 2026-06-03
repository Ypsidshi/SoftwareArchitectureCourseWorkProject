import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import path from "node:path";

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  server: {
    port: 5173,
    proxy: {
      "/api/auth": { target: "http://localhost:8082", changeOrigin: true },
      "/api/sanatoriums": { target: "http://localhost:8082", changeOrigin: true },
      "/api/medical-profiles": { target: "http://localhost:8082", changeOrigin: true },
      "/api/bookings": { target: "http://localhost:8082", changeOrigin: true },
      "/api/admin": { target: "http://localhost:8082", changeOrigin: true },
    },
  },
});
