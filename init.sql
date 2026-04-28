CREATE TABLE IF NOT EXISTS short_urls (
    code text NOT NULL,
    original_url text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT short_urls_pkey PRIMARY KEY (code),
    CONSTRAINT short_urls_original_url_key UNIQUE (original_url)
);
