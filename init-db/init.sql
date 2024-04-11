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
        token VARCHAR(255) UNIQUE NOT NULL,
        is_admin BOOLEAN NOT NULL DEFAULT false
    );

    CREATE TABLE public.features (
        feature_id SERIAL PRIMARY KEY,
        name VARCHAR(255) NOT NULL
    );