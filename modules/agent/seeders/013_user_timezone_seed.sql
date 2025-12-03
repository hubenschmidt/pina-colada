-- Seeder 013: User Timezone Preferences
-- Sets timezone preferences for existing users
-- Runs after 001_initial_seed.sql (creates William, Jennifer) and migration 040 (creates AI agent)

DO $$
DECLARE
    v_user_id BIGINT;
    v_prefs_id BIGINT;
BEGIN
    -- Set timezone for William Hubenschmidt (America/New_York)
    SELECT id INTO v_user_id FROM "User" WHERE email = 'whubenschmidt@gmail.com' LIMIT 1;
    IF v_user_id IS NOT NULL THEN
        INSERT INTO "User_Preferences" (user_id, timezone, created_at, updated_at)
        VALUES (v_user_id, 'America/New_York', NOW(), NOW())
        ON CONFLICT (user_id) DO UPDATE SET timezone = 'America/New_York', updated_at = NOW()
        RETURNING id INTO v_prefs_id;
        RAISE NOTICE 'Set timezone America/New_York for William Hubenschmidt (user_id: %)', v_user_id;
    END IF;

    -- Set timezone for Jennifer Lev (America/New_York)
    SELECT id INTO v_user_id FROM "User" WHERE email = 'jennifervlev@gmail.com' LIMIT 1;
    IF v_user_id IS NOT NULL THEN
        INSERT INTO "User_Preferences" (user_id, timezone, created_at, updated_at)
        VALUES (v_user_id, 'America/New_York', NOW(), NOW())
        ON CONFLICT (user_id) DO UPDATE SET timezone = 'America/New_York', updated_at = NOW()
        RETURNING id INTO v_prefs_id;
        RAISE NOTICE 'Set timezone America/New_York for Jennifer Lev (user_id: %)', v_user_id;
    END IF;

    -- Set timezone for AI Agent users (UTC) - all system users
    FOR v_user_id IN SELECT id FROM "User" WHERE is_system_user = TRUE
    LOOP
        INSERT INTO "User_Preferences" (user_id, timezone, created_at, updated_at)
        VALUES (v_user_id, 'UTC', NOW(), NOW())
        ON CONFLICT (user_id) DO UPDATE SET timezone = 'UTC', updated_at = NOW()
        RETURNING id INTO v_prefs_id;
        RAISE NOTICE 'Set timezone UTC for AI Agent (user_id: %)', v_user_id;
    END LOOP;

    RAISE NOTICE 'User timezone preferences seeded successfully';
END $$;

-- Return count for seeder runner
SELECT COUNT(*) AS users_with_timezone FROM "User_Preferences" WHERE timezone IS NOT NULL;
