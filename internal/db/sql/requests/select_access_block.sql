SELECT 
    owner, 
    COALESCE(access, '') 
FROM 
    files 
LEFT JOIN 
    file_access 
USING(file_id) 
WHERE 
    file_id=(SELECT file_id FROM blocks WHERE block_id=$1)
AND 
    (user_id=$2 OR owner=$2);