CREATE TABLE public.banners (
    banner_id SERIAL PRIMARY KEY,
    feature_id INTEGER NOT NULL,
    content JSONB NOT NULL,
    is_active BOOLEAN NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE public.tags (
    tag_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

CREATE TABLE public.banner_tags (
    banner_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    PRIMARY KEY (banner_id, tag_id),
    FOREIGN KEY (banner_id) REFERENCES banners(banner_id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(tag_id) ON DELETE CASCADE
);

CREATE TABLE public.users (
    user_id SERIAL PRIMARY KEY,
    token VARCHAR(255),
    is_admin BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE public.features (
    feature_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

INSERT INTO public.tags (name)
SELECT 'Tag ' || generate_series FROM generate_series(1, 400);

INSERT INTO public.features (name)
SELECT 'Feature ' || generate_series FROM generate_series(1, 400);

INSERT INTO public.users (is_admin)
SELECT false FROM generate_series(1, 20)
UNION ALL
SELECT true FROM generate_series(1, 5);

INSERT INTO public.banners (feature_id, content, is_active)
SELECT
    (random() * (SELECT max(feature_id) FROM public.features))::int + 1,
    jsonb_build_object('title', 'Banner ' || generate_series, 'description', md5(random()::text)),
    (random() > 0.5)
FROM generate_series(1, 1000);

DO $$
DECLARE
    banner_id int;
    tag_id_max int;
    attempts int;
BEGIN
    SELECT max(tag_id) INTO tag_id_max FROM public.tags;
    
    FOR banner_id IN 1..1000 LOOP
        attempts := (random() * 4 + 1)::int;
        WHILE attempts > 0 LOOP
            BEGIN
                INSERT INTO public.banner_tags (banner_id, tag_id)
                VALUES (banner_id, (random() * (tag_id_max - 1) + 1)::int)
                ON CONFLICT DO NOTHING;
                
                attempts := attempts - 1;
            EXCEPTION WHEN unique_violation THEN
                NULL;
            END;
        END LOOP;
    END LOOP;
END $$;

