import { createApp, h } from 'vue';
import { createInertiaApp } from '@inertiajs/vue3';
import './styles/app.css';

void createInertiaApp({
	title: (title) => (title ? `${title} - Alloy Inertia Demo` : 'Alloy Inertia Demo'),
	defaults: {
		visitOptions: (_href, options) => ({
			...options,
			preserveScroll: options?.preserveScroll ?? 'errors',
		}),
	},
	resolve: (name) => {
		const pages = import.meta.glob('./pages/**/*.vue');
		return pages[`./pages/${name}.vue`]();
	},
	setup({ el, App, props, plugin }) {
		createApp({ render: () => h(App, props) })
			.use(plugin)
			.mount(el);
	},
});
