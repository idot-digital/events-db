-- Create events table if it doesn't exist
CREATE TABLE IF NOT EXISTS events (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    source VARCHAR(255) NOT NULL,
    type VARCHAR(255) NOT NULL,
    subject VARCHAR(255) NOT NULL,
    time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    data VARBINARY(60000) NOT NULL,
    INDEX idx_subject_id (subject, id),
    INDEX idx_subject_type_id (subject, type, id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;