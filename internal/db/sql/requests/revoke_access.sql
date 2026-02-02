DELETE FROM
    file_access
WHERE
    file_id = $1
AND
    user_id = $2;