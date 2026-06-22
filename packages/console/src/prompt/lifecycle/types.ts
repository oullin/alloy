export type PromptReader<T> = (attempt: number) => Promise<T>;

export type PromptValidHandler<T> = (value: T) => void | Promise<void>;

export type PromptInvalidHandler<T> = (value?: T) => void | Promise<void>;
