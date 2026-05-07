ALTER TABLE stock_transfers 
RENAME COLUMN from_store_id TO from_location_id,
RENAME COLUMN to_store_id TO to_location_id,
RENAME COLUMN transfer_date TO transaction_date,
ADD COLUMN ref_no VARCHAR(50) NOT NULL AFTER to_location_id,
ADD COLUMN final_total DECIMAL(15,2) DEFAULT 0.00 AFTER status;

-- Drop old foreign keys and add new ones to business_locations
ALTER TABLE stock_transfers DROP FOREIGN KEY stock_transfers_ibfk_2;
ALTER TABLE stock_transfers DROP FOREIGN KEY stock_transfers_ibfk_3;

ALTER TABLE stock_transfers 
ADD CONSTRAINT fk_st_from_location FOREIGN KEY (from_location_id) REFERENCES business_locations(id) ON DELETE CASCADE,
ADD CONSTRAINT fk_st_to_location FOREIGN KEY (to_location_id) REFERENCES business_locations(id) ON DELETE CASCADE;
