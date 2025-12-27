-- Test data setup script for trade service integration tests
-- This script creates the required test user and sets up initial data

USE aether_defense;

-- Create test user (ID 1001) if not exists
-- Using INSERT IGNORE to avoid errors if user already exists
INSERT IGNORE INTO user (id, username, mobile, status, create_time, update_time)
VALUES (1001, 'testuser', '13800138000', 1, NOW(), NOW());

-- Verify the user was created
SELECT 'Test user created successfully' AS status, id, username, mobile, status 
FROM user 
WHERE id = 1001;

