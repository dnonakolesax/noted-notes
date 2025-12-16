UPDATE
    blocks
SET
    prev_id=$2
WHERE
    block_id = $1;   
