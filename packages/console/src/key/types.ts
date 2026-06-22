import type { Key } from '#console/key/constants';

export type KeyValue = string | readonly string[];

export type KeyName = Extract<(typeof Key)[keyof typeof Key], string>;

export type KeyboardEventLike = {
	name?: string;
	sequence?: string;
	ctrl?: boolean;
	meta?: boolean;
	shift?: boolean;
};
