UPDATE
    blocks
SET
    prev_id=$1
WHERE
    block_id = (SELECT block_id FROM blocks WHERE prev_id=$2 AND block_id <> $1);   

