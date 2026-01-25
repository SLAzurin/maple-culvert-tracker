import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// https://vite.dev/config/
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
});
