-- Add qrcode_image column to bots table
ALTER TABLE bots ADD COLUMN IF NOT EXISTS qrcode_image TEXT;
