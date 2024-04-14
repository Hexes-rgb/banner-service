CREATE TABLE public.features (
    feature_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

CREATE TABLE public.tags (
    tag_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

CREATE TABLE public.banners (
    banner_id SERIAL PRIMARY KEY,
    feature_id INTEGER NOT NULl,
    content JSONB NOT NULL,
    is_active BOOLEAN NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (feature_id) REFERENCES public.features(feature_id) ON DELETE CASCADE
);

CREATE TABLE public.users (
    user_id SERIAL PRIMARY KEY,
    token VARCHAR(255),
    is_admin BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE public.banner_tag (
    banner_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    FOREIGN KEY (banner_id) REFERENCES public.banners(banner_id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES public.tags(tag_id) ON DELETE CASCADE,
    PRIMARY KEY (banner_id, tag_id)
);

CREATE OR REPLACE FUNCTION check_unique_feature_tag_combination()
RETURNS TRIGGER AS $$
DECLARE
    v_banner_feature_id INTEGER;
BEGIN
    SELECT feature_id INTO v_banner_feature_id FROM banners WHERE banner_id = NEW.banner_id;

    IF (SELECT COUNT(*) FROM banner_tag
        JOIN banners ON banners.banner_id = banner_tag.banner_id
        WHERE banners.feature_id = v_banner_feature_id
        AND banner_tag.tag_id = NEW.tag_id
        AND banner_tag.banner_id <> NEW.banner_id) > 0 THEN
        RAISE EXCEPTION 'Not a unique combination of tag_id and feature_id';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_check_unique_feature_tag_combination
BEFORE INSERT OR UPDATE ON banner_tag
FOR EACH ROW EXECUTE FUNCTION check_unique_feature_tag_combination();

INSERT INTO public.features (name)
SELECT 'Feature ' || i
FROM generate_series(1, 500) AS i;

INSERT INTO public.tags (name)
SELECT 'Tag ' || i
FROM generate_series(1, 800) AS i;

INSERT INTO public.users (token, is_admin)
SELECT '', i <= 3
FROM generate_series(1, 10) AS i;

DO $$
DECLARE
    total_features INT;
    random_feature_id INT;
BEGIN
    SELECT COUNT(*) INTO total_features FROM public.features;

    FOR i IN 1..3000 LOOP
        random_feature_id := trunc(random() * (total_features - 1)) + 1;

        INSERT INTO public.banners (feature_id, content, is_active)
        VALUES (random_feature_id, '{}'::jsonb, true);
    END LOOP;
END $$;


DO $$
DECLARE
    v_banner_id INTEGER;
    v_tag_id INTEGER;
    v_count_tags INTEGER;
    v_total_tags INTEGER;
BEGIN
    SELECT COUNT(*) INTO v_total_tags FROM public.tags;

    FOR v_banner_id IN SELECT banner_id FROM public.banners LOOP
        v_count_tags := trunc(random() * 7) + 1;

        WHILE v_count_tags > 0 LOOP
            v_tag_id := trunc(random() * (v_total_tags-1)) + 1;
            BEGIN
                INSERT INTO public.banner_tag (banner_id, tag_id) VALUES (v_banner_id, v_tag_id);
                v_count_tags := v_count_tags - 1;
            EXCEPTION WHEN OTHERS THEN
                IF SQLSTATE = '45000' THEN
                    CONTINUE;
                ELSE
                    RAISE NOTICE 'Error: %', SQLERRM;
                END IF;
            END;
        END LOOP;
    END LOOP;
END $$;



