UPDATE 
    blocks
SET parent_id = (SELECT parent_id FROM blocks WHERE block_id=$1)
WHERE 
    parent_id = $1;

