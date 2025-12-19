UPDATE 
    blocks
SET prev_id = (SELECT prev_id FROM blocks WHERE block_id=$1)
WHERE 
    prev_id = $1;

