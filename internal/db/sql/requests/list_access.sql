SELECT
    user_id, access
FROM
    file_access
WHERE 
    file_id=$1;