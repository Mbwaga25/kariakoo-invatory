ALTER TABLE business_locations 
ADD COLUMN location_id VARCHAR(100) AFTER name,
ADD COLUMN state VARCHAR(100) AFTER city,
ADD COLUMN country VARCHAR(100) AFTER state,
ADD COLUMN zip_code VARCHAR(20) AFTER country;
