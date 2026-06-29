import { z } from 'zod';
import { tableCellArraySchema, tableCellRecordSchema } from '#console/output/validators/cells';
import type { DataTableObjectRow } from '#console/types';

export const dataTableArrayRowSchema = tableCellArraySchema;

export const dataTableRecordRowSchema = tableCellRecordSchema;

export const dataTableStepNameSchema = z.string();

export const dataObjectRowSchema: z.ZodType<DataTableObjectRow<unknown>> = z
	.object({
		cells: tableCellRecordSchema,
		value: z.unknown().optional(),
	})
	.passthrough();

export const dataTableRowSchema = z.union([dataObjectRowSchema, dataTableArrayRowSchema, dataTableRecordRowSchema]);

export const dataTableRowsSchema = z.array(dataTableRowSchema).min(1, 'Data table must have at least one row.');

export const dataTablePromptOptionsSchema = z
	.object({
		headers: z.array(z.string()).optional(),
		message: z.string(),
		rows: dataTableRowsSchema,
		scroll: z.number().optional(),
	})
	.passthrough();

export const dataTableHeadersSchema = z.array(z.string());
