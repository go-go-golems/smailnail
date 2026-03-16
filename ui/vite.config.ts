import react from "@vitejs/plugin-react";
import { defineConfig, loadEnv } from "vite";
import path from "node:path";

function parsePort(rawValue: string | undefined, fallback: number): number {
  if (!rawValue) {
    return fallback;
  }

  const parsed = Number.parseInt(rawValue, 10);
  if (Number.isNaN(parsed) || parsed <= 0) {
    return fallback;
  }

  return parsed;
}

function resolveBackendTarget(env: Record<string, string>): string {
  const explicitTarget = env.SMAILNAIL_UI_BACKEND_URL?.trim();
  if (explicitTarget) {
    return explicitTarget;
  }

  const protocol = env.SMAILNAIL_UI_BACKEND_PROTOCOL?.trim() || "http";
  const host = env.SMAILNAIL_UI_BACKEND_HOST?.trim() || "localhost";
  const port = parsePort(env.SMAILNAIL_UI_BACKEND_PORT, 8080);

  return `${protocol}://${host}:${port}`;
}

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");
  const devPort = parsePort(env.SMAILNAIL_UI_DEV_PORT, 5050);
  const backendTarget = resolveBackendTarget(env);

  return {
    plugins: [react()],
    build: {
      outDir: path.resolve(import.meta.dirname, "dist/public"),
      emptyOutDir: true,
    },
    resolve: {
      alias: {
        "@": path.resolve(import.meta.dirname, "src"),
      },
    },
    server: {
      port: devPort,
      strictPort: false,
      host: true,
      proxy: {
        "/api": {
          target: backendTarget,
          changeOrigin: true,
        },
      },
    },
  };
});
