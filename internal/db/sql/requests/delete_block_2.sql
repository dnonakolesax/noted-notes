DELETE FROM
    blocks
WHERE 
    block_id=$1  
AND
    file_id=$2;    
