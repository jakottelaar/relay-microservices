DROP TRIGGER IF EXISTS update_accounts_updated_at ON accounts;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP INDEX IF EXISTS idx_accounts_email;
DROP TABLE IF EXISTS accounts;