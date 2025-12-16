DELETE FROM blocks
WHERE file_id=$1
RETURNING block_id;