SELECT 
    f.filename,
    f.public,
    COALESCE(ARRAY_AGG(b.block_id) FILTER (WHERE b.block_id IS NOT NULL), '{}') AS blocks,
    COALESCE(ARRAY_AGG(b.language) FILTER (WHERE b.block_id IS NOT NULL), '{}') AS languages
FROM 
    files f
LEFT JOIN 
    blocks b ON f.file_id = b.file_id
WHERE 
    f.file_id = $1
GROUP BY 
    f.file_id;