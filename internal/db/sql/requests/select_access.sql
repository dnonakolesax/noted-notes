SELECT 
    owner, 
    COALESCE(access, ''),
    public
FROM 
    files 
LEFT JOIN 
    file_access 
USING(file_id) 
WHERE 
    file_id=$1 
AND 
    (user_id=$2 OR owner=$2 or public=true);