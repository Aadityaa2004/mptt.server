-- Add sap_alert_threshold_percent to users table (0-100, null = use default from config)
ALTER TABLE users ADD COLUMN IF NOT EXISTS sap_alert_threshold_percent INTEGER;

COMMENT ON COLUMN users.sap_alert_threshold_percent IS 'Sap fill percentage (0-100) at which to send email alert. NULL uses system default.';
