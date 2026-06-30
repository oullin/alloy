import { fileURLToPath } from 'node:url';
import { defineConfig } from 'vite-plus';
import vue from '@vitejs/plugin-vue';
import tailwindcss from '@tailwindcss/vite';

const appPath = (value) => fileURLToPath(new URL(value, import.meta.url));
const repoPath = (value) => fileURLToPath(new URL(`../../../${value}`, import.meta.url));

export default defineConfig({
	plugins: [vue(), tailwindcss()],
	base: '/assets/',
	root: appPath('./'),
	resolve: {
		alias: {
			'@': appPath('src'),
		},
	},
	build: {
		outDir: repoPath('storage/apps/inertia/dist/app'),
		emptyOutDir: true,
		manifest: true,
		rollupOptions: {
			input: appPath('src/js/app.js'),
			output: {
				entryFileNames: 'app.js',
				chunkFileNames: '[name]-[hash].js',
				assetFileNames: 'app.[ext]',
			},
		},
	},
});
