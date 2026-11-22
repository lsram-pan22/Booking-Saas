BEGIN;
ALTER TABLE subscriptions ADD COLUMN selection_required BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE tenants ADD COLUMN plan_enabled BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE tenant_members ADD COLUMN plan_enabled BOOLEAN NOT NULL DEFAULT TRUE;
-- Retain expired subscriptions as history instead of enforcing one row per user.
DO $$
DECLARE item RECORD; user_column SMALLINT;
BEGIN
 SELECT attnum INTO user_column FROM pg_attribute WHERE attrelid='subscriptions'::regclass AND attname='user_id';
 FOR item IN SELECT conname FROM pg_constraint WHERE conrelid='subscriptions'::regclass AND contype='u' AND conkey=ARRAY[user_column] LOOP
  EXECUTE format('ALTER TABLE subscriptions DROP CONSTRAINT %I',item.conname);
 END LOOP;
 FOR item IN SELECT i.indexrelid::regclass AS index_name FROM pg_index i WHERE i.indrelid='subscriptions'::regclass AND i.indisunique AND NOT i.indisprimary AND i.indpred IS NULL AND i.indnkeyatts=1 AND i.indkey[0]=user_column AND NOT EXISTS(SELECT 1 FROM pg_constraint c WHERE c.conindid=i.indexrelid) LOOP
  EXECUTE format('DROP INDEX %s',item.index_name);
 END LOOP;
END; $$;
CREATE UNIQUE INDEX subscriptions_one_active_user ON subscriptions(user_id) WHERE status='active';
CREATE INDEX subscriptions_user_history ON subscriptions(user_id,starts_at DESC,id DESC);
COMMIT;
