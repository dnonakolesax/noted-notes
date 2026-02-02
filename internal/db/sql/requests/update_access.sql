UPDATE 
    file_access 
SET
    access=$3
WHERE
    file_id = $1
AND
    user_id = $2;