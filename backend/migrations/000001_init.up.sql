-- Task Management Application - Initial Schema Migration
-- Creates the foundational tables: users, tasks, and activity_logs
-- with proper constraints, indexes, and performance optimizations.

BEGIN;

-- #############################################################################
-- TABLE: users
-- Stores registered user accounts with role-based access control.
-- #############################################################################
CREATE TABLE IF NOT EXISTS users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL,
    role        VARCHAR(20)  NOT NULL DEFAULT 'user'
                CHECK (role IN ('user', 'admin')),
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_users_role  ON users (role);

-- #############################################################################
-- TABLE: tasks
-- Core task entity linked to a user. Supports status and priority tracking.
-- #############################################################################
CREATE TABLE IF NOT EXISTS tasks (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title       VARCHAR(255) NOT NULL,
    description TEXT        DEFAULT '',
    status      VARCHAR(20) NOT NULL DEFAULT 'pending'
                CHECK (status IN ('pending', 'in_progress', 'completed', 'cancelled')),
    priority    VARCHAR(10) NOT NULL DEFAULT 'medium'
                CHECK (priority IN ('low', 'medium', 'high', 'critical')),
    due_date    TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Performance indexes for common query filters.
CREATE INDEX idx_tasks_user_id      ON tasks (user_id);
CREATE INDEX idx_tasks_status       ON tasks (status);
CREATE INDEX idx_tasks_due_date     ON tasks (due_date);

-- Composite index for filtering tasks by user + status (most common query).
CREATE INDEX idx_tasks_user_status  ON tasks (user_id, status);

-- Composite index for sorting by due date within a user's tasks.
CREATE INDEX idx_tasks_user_duedate ON tasks (user_id, due_date);

-- #############################################################################
-- TABLE: activity_logs
-- Immutable audit log tracking all task-related actions.
-- #############################################################################
CREATE TABLE IF NOT EXISTS activity_logs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id     UUID        NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action      VARCHAR(50) NOT NULL
                CHECK (action IN (
                    'task_created', 'task_updated', 'task_deleted',
                    'status_changed', 'priority_changed', 'assigned',
                    'comment_added', 'due_date_changed'
                )),
    details     JSONB       DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_activity_logs_task_id  ON activity_logs (task_id);
CREATE INDEX idx_activity_logs_user_id  ON activity_logs (user_id);
CREATE INDEX idx_activity_logs_action   ON activity_logs (action);
CREATE INDEX idx_activity_logs_created  ON activity_logs (created_at DESC);

-- Composite index for fetching a task's activity timeline efficiently.
CREATE INDEX idx_activity_logs_task_created ON activity_logs (task_id, created_at DESC);

COMMIT;