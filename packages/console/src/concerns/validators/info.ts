import { z } from 'zod';
import type { InfoResolver } from '#console/concerns/info';

type InfoCallback<T> = Exclude<InfoResolver<T>, string | undefined>;

const infoCallbackSchema = z.function();

export const isInfoCallback = <T>(info: InfoResolver<T>): info is InfoCallback<T> => {
	return infoCallbackSchema.safeParse(info).success;
};
