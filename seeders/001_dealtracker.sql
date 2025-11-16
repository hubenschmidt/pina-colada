-- DealTracker Seed Data
-- Seeds the Status table and configures workflows for Jobs, Leads, Deals, Tasks

-- ==============================
-- 1. Insert Status records
-- ==============================

-- Job Statuses
INSERT INTO "Status" (name, category, is_terminal, description) VALUES
('Applied', 'job', false, 'Job application submitted'),
('Interviewing', 'job', false, 'In interview process'),
('Offer', 'job', false, 'Offer received'),
('Accepted', 'job', true, 'Job offer accepted'),
('Rejected', 'job', true, 'Application or candidacy rejected');

-- Lead Stages
INSERT INTO "Status" (name, category, is_terminal, description) VALUES
('New', 'lead', false, 'New lead created'),
('Qualified', 'lead', false, 'Lead has been qualified'),
('Discovery', 'lead', false, 'Discovery phase'),
('Proposal', 'lead', false, 'Proposal submitted'),
('Negotiation', 'lead', false, 'In negotiation'),
('Won', 'lead', true, 'Lead won'),
('Lost', 'lead', true, 'Lead lost'),
('On Hold', 'lead', false, 'Lead on hold');

-- Deal Stages
INSERT INTO "Status" (name, category, is_terminal, description) VALUES
('Prospecting', 'deal', false, 'Prospecting phase'),
('Qualification', 'deal', false, 'Qualification phase'),
-- Note: 'Proposal' and 'Negotiation' already exist from Lead stages
('Closed Won', 'deal', true, 'Deal won'),
('Closed Lost', 'deal', true, 'Deal lost');

-- Task Statuses
INSERT INTO "Status" (name, category, is_terminal, description) VALUES
('To Do', 'task_status', false, 'Task not started'),
('In Progress', 'task_status', false, 'Task in progress'),
('Completed', 'task_status', true, 'Task completed'),
('Cancelled', 'task_status', true, 'Task cancelled');

-- Task Priorities
INSERT INTO "Status" (name, category, is_terminal, description) VALUES
('Low', 'task_priority', false, 'Low priority'),
('Medium', 'task_priority', false, 'Medium priority'),
('High', 'task_priority', false, 'High priority'),
('Urgent', 'task_priority', false, 'Urgent priority');

-- ==============================
-- 2. Configure Job workflow
-- ==============================

INSERT INTO "Job_Status" (status_id, display_order, is_default)
SELECT id, 0, true FROM "Status" WHERE name = 'Applied' AND category = 'job'
UNION ALL
SELECT id, 1, false FROM "Status" WHERE name = 'Interviewing' AND category = 'job'
UNION ALL
SELECT id, 2, false FROM "Status" WHERE name = 'Offer' AND category = 'job'
UNION ALL
SELECT id, 3, false FROM "Status" WHERE name = 'Accepted' AND category = 'job'
UNION ALL
SELECT id, 4, false FROM "Status" WHERE name = 'Rejected' AND category = 'job';

-- ==============================
-- 3. Configure Lead workflow
-- ==============================

INSERT INTO "Lead_Status" (status_id, display_order, is_default)
SELECT id, 0, true FROM "Status" WHERE name = 'New' AND category = 'lead'
UNION ALL
SELECT id, 1, false FROM "Status" WHERE name = 'Qualified' AND category = 'lead'
UNION ALL
SELECT id, 2, false FROM "Status" WHERE name = 'Discovery' AND category = 'lead'
UNION ALL
SELECT id, 3, false FROM "Status" WHERE name = 'Proposal' AND category = 'lead'
UNION ALL
SELECT id, 4, false FROM "Status" WHERE name = 'Negotiation' AND category = 'lead'
UNION ALL
SELECT id, 5, false FROM "Status" WHERE name = 'Won' AND category = 'lead'
UNION ALL
SELECT id, 6, false FROM "Status" WHERE name = 'Lost' AND category = 'lead'
UNION ALL
SELECT id, 7, false FROM "Status" WHERE name = 'On Hold' AND category = 'lead';

-- ==============================
-- 4. Configure Deal workflow
-- ==============================

INSERT INTO "Deal_Status" (status_id, display_order, is_default)
SELECT id, 0, true FROM "Status" WHERE name = 'Prospecting' AND category = 'deal'
UNION ALL
SELECT id, 1, false FROM "Status" WHERE name = 'Qualification' AND category = 'deal'
UNION ALL
SELECT id, 2, false FROM "Status" WHERE name = 'Proposal' AND category = 'lead'  -- Reusing from Lead
UNION ALL
SELECT id, 3, false FROM "Status" WHERE name = 'Negotiation' AND category = 'lead'  -- Reusing from Lead
UNION ALL
SELECT id, 4, false FROM "Status" WHERE name = 'Closed Won' AND category = 'deal'
UNION ALL
SELECT id, 5, false FROM "Status" WHERE name = 'Closed Lost' AND category = 'deal';

-- ==============================
-- 5. Configure Task statuses
-- ==============================

INSERT INTO "Task_Status" (status_id, display_order, is_default)
SELECT id, 0, true FROM "Status" WHERE name = 'To Do' AND category = 'task_status'
UNION ALL
SELECT id, 1, false FROM "Status" WHERE name = 'In Progress' AND category = 'task_status'
UNION ALL
SELECT id, 2, false FROM "Status" WHERE name = 'Completed' AND category = 'task_status'
UNION ALL
SELECT id, 3, false FROM "Status" WHERE name = 'Cancelled' AND category = 'task_status';

-- ==============================
-- 6. Configure Task priorities
-- ==============================

INSERT INTO "Task_Priority" (status_id, display_order, is_default)
SELECT id, 0, false FROM "Status" WHERE name = 'Low' AND category = 'task_priority'
UNION ALL
SELECT id, 1, true FROM "Status" WHERE name = 'Medium' AND category = 'task_priority'  -- Medium is default
UNION ALL
SELECT id, 2, false FROM "Status" WHERE name = 'High' AND category = 'task_priority'
UNION ALL
SELECT id, 3, false FROM "Status" WHERE name = 'Urgent' AND category = 'task_priority';
