SELECT 
    COUNT(*) 
FROM 
    blocks 
WHERE 
    ((block_id=$1 AND prev_id=$2)
    OR 
    block_id=$2) 
AND 
    file_id=$3;