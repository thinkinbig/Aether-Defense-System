-- Grant access to aether user from any host
-- This allows connections from Docker host (172.19.0.1) and other containers
-- Note: The user 'aether'@'localhost' is already created by MYSQL_USER environment variable
-- and already has privileges. This script creates 'aether'@'%' for remote access.

-- Create user for '%' host if it doesn't exist (MySQL 8.0 compatible)
-- The user 'aether'@'localhost' already exists from MYSQL_USER env var with full privileges
-- We need to create 'aether'@'%' separately for remote access
-- Note: CREATE USER IF NOT EXISTS is supported in MySQL 8.0.11+
CREATE USER IF NOT EXISTS 'aether'@'%' IDENTIFIED BY 'aether123';

-- Grant all privileges on the database to '%' host (for remote access)
-- Note: 'aether'@'localhost' already has privileges from the entrypoint script
GRANT ALL PRIVILEGES ON aether_defense.* TO 'aether'@'%';

-- Flush privileges to apply changes
FLUSH PRIVILEGES;

