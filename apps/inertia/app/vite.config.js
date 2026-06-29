import path from 'node:path';
import { defineConfig } from 'vite-plus';
import vue from '@vitejs/plugin-vue';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
	plugins: [vue(), tailwindcss()],
	base: '/assets/',
	root: '.',
	resolve: {
		alias: {
			'@': path.resolve(__dirname, 'src'),
		},
	},
	build: {
		outDir: '../../../storage/apps/inertia/dist/app',
		emptyOutDir: true,
		manifest: true,
		rollupOptions: {
			input: 'src/js/app.js',
			output: {
				entryFileNames: 'app.js',
				chunkFileNames: '[name]-[hash].js',
				assetFileNames: 'app.[ext]',
			},
		},
	},
});
