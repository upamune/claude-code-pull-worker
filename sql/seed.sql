-- Insert default global settings
INSERT OR IGNORE INTO global_settings (setting_key, setting_value) VALUES 
    ('default_notification_config', '{"discord": {"webhook_url": ""}}');