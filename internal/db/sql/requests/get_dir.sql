SELECT
    file_id,
    filename,
    dir,
	COALESCE(access, '') as access,
	owner
FROM
    files
    LEFT JOIN file_access USING (file_id)
WHERE
    ((
        public = true
    ) OR (
		access != null
        AND
        file_access.user_id = $2
	) OR (
		owner=$2
	))
    AND parent_id = $1;