import { z } from 'zod';

const callableSchema = z.function();

export const functionSchema = <T>(): z.ZodType<T> => z.custom<T>((value) => callableSchema.safeParse(value).success);
