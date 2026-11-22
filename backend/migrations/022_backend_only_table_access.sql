BEGIN;
-- Application JWTs are verified by the Go API, not by Supabase's Data API.
-- No browser role receives direct access to these tables.
DO $$
DECLARE table_name TEXT; role_name TEXT;
BEGIN
 FOR table_name IN SELECT unnest(ARRAY['users','packages','subscriptions','tenants','tenant_members','services','bookings','booking_status_history','booking_edit_history','password_events','schema_migrations']) LOOP
  EXECUTE format('ALTER TABLE public.%I ENABLE ROW LEVEL SECURITY', table_name);
  EXECUTE format('REVOKE ALL ON TABLE public.%I FROM PUBLIC', table_name);
  FOR role_name IN SELECT rolname FROM pg_roles WHERE rolname IN ('anon','authenticated') LOOP
   EXECUTE format('REVOKE ALL ON TABLE public.%I FROM %I', table_name, role_name);
  END LOOP;
 END LOOP;
END; $$;
COMMIT;
