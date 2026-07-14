export class StreamRenderScheduler {
	#scheduled = false;

	flush(render: () => void): void {
		if (!this.#scheduled) {
			render();

			return;
		}

		this.#scheduled = false;
		render();
	}

	request(render: () => void): void {
		if (this.#scheduled) {
			return;
		}

		this.#scheduled = true;
		queueMicrotask(() => {
			if (!this.#scheduled) {
				return;
			}

			this.#scheduled = false;
			render();
		});
	}
}
