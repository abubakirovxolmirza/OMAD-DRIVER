-- Remove all duplicate regions and districts
-- Keep only the original 14 regions (ID 1-14) and 191 districts

-- Delete all duplicate regions (created by multiple migrations)
DELETE FROM regions WHERE id > 14;

-- Reset sequences to correct values
SELECT setval('regions_id_seq', 14);
SELECT setval('districts_id_seq', (SELECT MAX(id) FROM districts));

-- Verify results
SELECT COUNT(*) as total_regions FROM regions;
SELECT COUNT(*) as total_districts FROM districts;

