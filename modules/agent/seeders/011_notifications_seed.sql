-- Seeder 011: Comment Notifications
-- Creates notifications for existing comments to demonstrate the notification system
-- Notifications are created for comments made by the "other" user to simulate realistic activity

DO $$
DECLARE
    v_tenant_id BIGINT;
    v_user_id_william BIGINT;
    v_user_id_jennifer BIGINT;
    v_notifications_inserted INTEGER := 0;
    v_comment RECORD;
BEGIN
    -- Get the tenant ID
    SELECT id INTO v_tenant_id FROM "Tenant" WHERE slug = 'pinacolada' LIMIT 1;

    IF v_tenant_id IS NULL THEN
        RAISE NOTICE 'No tenant found, skipping notifications seeder';
        RETURN;
    END IF;

    -- Get user IDs
    SELECT id INTO v_user_id_william FROM "User" WHERE email = 'whubenschmidt@gmail.com' AND tenant_id = v_tenant_id LIMIT 1;
    SELECT id INTO v_user_id_jennifer FROM "User" WHERE email = 'jennifervlev@gmail.com' AND tenant_id = v_tenant_id LIMIT 1;

    IF v_user_id_william IS NULL OR v_user_id_jennifer IS NULL THEN
        RAISE NOTICE 'Users not found, skipping notifications seeder';
        RETURN;
    END IF;

    -- Clear existing notifications for clean seeding
    DELETE FROM "CommentNotification" WHERE tenant_id = v_tenant_id;

    -- Create notifications for William from Jennifer's comments (thread_activity)
    -- These simulate Jennifer commenting on threads where William has also commented
    INSERT INTO "CommentNotification" (tenant_id, user_id, comment_id, notification_type, is_read, created_at)
    SELECT DISTINCT ON (c.id)
        v_tenant_id,
        v_user_id_william,
        c.id,
        'thread_activity',
        FALSE,
        c.created_at
    FROM "Comment" c
    WHERE c.tenant_id = v_tenant_id
      AND c.created_by = v_user_id_jennifer
      -- Only for entities where William has also commented
      AND EXISTS (
          SELECT 1 FROM "Comment" c2
          WHERE c2.commentable_type = c.commentable_type
            AND c2.commentable_id = c.commentable_id
            AND c2.created_by = v_user_id_william
      )
    ORDER BY c.id, c.created_at DESC
    LIMIT 8;

    GET DIAGNOSTICS v_notifications_inserted = ROW_COUNT;
    RAISE NOTICE 'Created % thread_activity notifications for William', v_notifications_inserted;

    -- Create notifications for Jennifer from William's comments (thread_activity)
    INSERT INTO "CommentNotification" (tenant_id, user_id, comment_id, notification_type, is_read, created_at)
    SELECT DISTINCT ON (c.id)
        v_tenant_id,
        v_user_id_jennifer,
        c.id,
        'thread_activity',
        FALSE,
        c.created_at
    FROM "Comment" c
    WHERE c.tenant_id = v_tenant_id
      AND c.created_by = v_user_id_william
      -- Only for entities where Jennifer has also commented
      AND EXISTS (
          SELECT 1 FROM "Comment" c2
          WHERE c2.commentable_type = c.commentable_type
            AND c2.commentable_id = c.commentable_id
            AND c2.created_by = v_user_id_jennifer
      )
    ORDER BY c.id, c.created_at DESC
    LIMIT 5;

    GET DIAGNOSTICS v_notifications_inserted = ROW_COUNT;
    RAISE NOTICE 'Created % thread_activity notifications for Jennifer', v_notifications_inserted;

    -- Create some direct_reply notifications for comments with parent_comment_id
    -- First, let's create a few reply comments if they don't exist
    INSERT INTO "Comment" (tenant_id, commentable_type, commentable_id, content, created_by, parent_comment_id, created_at, updated_at)
    SELECT
        v_tenant_id,
        c.commentable_type,
        c.commentable_id,
        CASE (ROW_NUMBER() OVER (ORDER BY c.id) % 3)
            WHEN 0 THEN 'Great point! I completely agree with this assessment.'
            WHEN 1 THEN 'Thanks for the update. Let me follow up on this tomorrow.'
            WHEN 2 THEN 'Good catch! I will make sure to address this in our next meeting.'
        END,
        CASE WHEN c.created_by = v_user_id_william THEN v_user_id_jennifer ELSE v_user_id_william END,
        c.id,
        c.created_at + INTERVAL '2 hours',
        c.created_at + INTERVAL '2 hours'
    FROM "Comment" c
    WHERE c.tenant_id = v_tenant_id
      AND c.parent_comment_id IS NULL
      AND c.created_by IN (v_user_id_william, v_user_id_jennifer)
    ORDER BY c.created_at DESC
    LIMIT 6
    ON CONFLICT DO NOTHING;

    GET DIAGNOSTICS v_notifications_inserted = ROW_COUNT;
    RAISE NOTICE 'Created % reply comments', v_notifications_inserted;

    -- Now create direct_reply notifications for these replies
    INSERT INTO "CommentNotification" (tenant_id, user_id, comment_id, notification_type, is_read, created_at)
    SELECT
        v_tenant_id,
        parent.created_by,  -- Notify the author of the parent comment
        reply.id,
        'direct_reply',
        FALSE,
        reply.created_at
    FROM "Comment" reply
    JOIN "Comment" parent ON reply.parent_comment_id = parent.id
    WHERE reply.tenant_id = v_tenant_id
      AND reply.created_by != parent.created_by  -- Don't notify yourself
      AND parent.created_by IS NOT NULL
    ON CONFLICT (user_id, comment_id) DO NOTHING;

    GET DIAGNOSTICS v_notifications_inserted = ROW_COUNT;
    RAISE NOTICE 'Created % direct_reply notifications', v_notifications_inserted;

    -- Mark some older notifications as read to show mixed state
    UPDATE "CommentNotification"
    SET is_read = TRUE, read_at = created_at + INTERVAL '1 day'
    WHERE tenant_id = v_tenant_id
      AND created_at < NOW() - INTERVAL '2 days'
      AND id IN (
          SELECT id FROM "CommentNotification"
          WHERE tenant_id = v_tenant_id
          ORDER BY RANDOM()
          LIMIT 3
      );

    RAISE NOTICE 'Notifications seeder completed successfully';
END $$;

-- Show notification counts for verification
SELECT
    u.email,
    COUNT(*) as total_notifications,
    SUM(CASE WHEN cn.is_read THEN 0 ELSE 1 END) as unread_count,
    SUM(CASE WHEN cn.notification_type = 'direct_reply' THEN 1 ELSE 0 END) as direct_replies,
    SUM(CASE WHEN cn.notification_type = 'thread_activity' THEN 1 ELSE 0 END) as thread_activity
FROM "CommentNotification" cn
JOIN "User" u ON cn.user_id = u.id
GROUP BY u.email
ORDER BY u.email;
