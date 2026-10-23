import path from "path";
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

const apiTarget = process.env.VITE_API_PROXY_TARGET || "http://localhost:8080";

const apiProxyConfig = {
  target: apiTarget,
  changeOrigin: true,
  bypass: (req: any) => {
    // Non-HTML requests (like direct file downloads or CSV export) should go to the backend
    if (req.url?.includes("/download") || req.url?.includes("/export")) {
      return undefined;
    }
    // If the browser is requesting an HTML document (navigation or page refresh), serve the SPA
    if (req.headers.accept?.includes("text/html")) {
      return "/index.html";
    }
    return undefined;
  },
};

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
      "/api": { target: apiTarget, changeOrigin: true },
      "/dashboard": apiProxyConfig,
      "/settings": apiProxyConfig,
      "/ebook": apiProxyConfig,
      "/file": apiProxyConfig,
      "/client": apiProxyConfig,
      "/purchase": apiProxyConfig,
      "/stripe-connect": apiProxyConfig,
      "/transactions": apiProxyConfig,
      "/sales": apiProxyConfig,
      "/checkout": apiProxyConfig,
      "/logout": { target: apiTarget, changeOrigin: true },
    },
  },
});
