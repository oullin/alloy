export const streamClosedError = (): Error => new Error('Stream is closed.');

export const streamPromptError = (): Error => new Error('Stream cannot be prompted');
