-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.files
(
    file_id uuid NOT NULL,
    filename character varying COLLATE pg_catalog."default" NOT NULL,
    dir boolean NOT NULL,
    parent_id uuid,
    public boolean,
    deleted_date timestamp with time zone,
    owner uuid,
    CONSTRAINT files_pkey PRIMARY KEY (file_id),
    CONSTRAINT files_filename_parent_id_key UNIQUE (filename, parent_id)
);

CREATE TABLE IF NOT EXISTS public.blocks
(
    block_id uuid NOT NULL,
    file_id uuid NOT NULL,
    prev_id uuid,
    language character varying COLLATE pg_catalog."default" NOT NULL
);

CREATE TABLE IF NOT EXISTS public.file_access
(
    file_id uuid NOT NULL,
    user_id uuid NOT NULL,
    access character varying COLLATE pg_catalog."default" NOT NULL,
    granted_by uuid,
    granted_date timestamp without time zone,
    CONSTRAINT file_access_pkey PRIMARY KEY (file_id, user_id),
    CONSTRAINT file FOREIGN KEY (file_id)
        REFERENCES public.files (file_id)
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS public.file_access
DROP TABLE IF EXISTS public.blocks
DROP TABLE IF EXISTS public.files
-- +goose StatementEnd
