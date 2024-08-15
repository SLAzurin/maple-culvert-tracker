import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";

// https://vitejs.dev/config/
export default defineConfig({
	plugins: [react()],
	server: {
		open: true,
		proxy: {
			"/api": {
				target: "http://localhost:" + (process.env.PORT || "8080"),
				changeOrigin: true,
				secure: false,
			},
		},
	},
	build: {
		outDir: "build",
		sourcemap: true,
	},
	test: {
		globals: true,
		environment: "jsdom",
		setupFiles: "src/setupTests",
		mockReset: true,
	},
});
