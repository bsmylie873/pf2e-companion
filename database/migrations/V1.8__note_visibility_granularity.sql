-- Migrate existing 'shared' notes to 'editable' (preserves current behavior)
UPDATE notes SET visibility = 'editable' WHERE visibility = 'shared';
