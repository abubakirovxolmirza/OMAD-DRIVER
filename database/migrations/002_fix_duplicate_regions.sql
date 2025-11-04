-- Remove duplicate or unwanted regions
-- Replace the IDs below with the actual IDs you want to delete

-- Example: Delete specific regions by ID
-- DELETE FROM regions WHERE id IN (15, 16, 17);

-- Example: Delete by name
-- DELETE FROM regions WHERE name_uz_lat = 'Duplicate Name';

-- To find duplicates first, run:
-- SELECT name_uz_lat, COUNT(*) FROM regions GROUP BY name_uz_lat HAVING COUNT(*) > 1;

-- Add your DELETE statements here:
