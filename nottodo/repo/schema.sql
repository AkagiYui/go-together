-- 缓存 ===============================
-- 使用 UNLOGGED 表以获得更好的性能
CREATE UNLOGGED TABLE IF NOT EXISTS app_cache (
    -- 缓存键，用于快速查找
    key VARCHAR(255) PRIMARY KEY,

    -- 缓存的值，使用 jsonb 类型。
    -- jsonb 是二进制格式，比 text 存储 json 效率更高，且支持索引和内部操作。
    value JSONB NOT NULL,

    -- 过期时间，TIMESTAMPTZ 存储带时区的时间，是最佳实践
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
ALTER TABLE app_cache OWNER TO nottodo;

-- 为过期时间创建一个索引，方便我们快速清理过期的缓存
CREATE INDEX IF NOT EXISTS idx_app_cache_expires_at ON app_cache (expires_at);

COMMENT ON TABLE app_cache IS '缓存';
COMMENT ON COLUMN app_cache.key IS '键';
COMMENT ON COLUMN app_cache.value IS '值';
COMMENT ON COLUMN app_cache.expires_at IS '过期时间';
COMMENT ON COLUMN app_cache.created_at IS '创建时间';


-- 系统设置 ===============================
CREATE TABLE IF NOT EXISTS settings (
    key VARCHAR(255) PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
ALTER TABLE settings OWNER TO nottodo;

COMMENT ON TABLE settings IS '系统设置';
COMMENT ON COLUMN settings.key IS '键';
COMMENT ON COLUMN settings.value IS '值';
COMMENT ON COLUMN settings.description IS '描述';
COMMENT ON COLUMN settings.updated_at IS '更新时间';


-- 待办事项 ===============================
CREATE TABLE IF NOT EXISTS todos (
	id BIGSERIAL PRIMARY KEY,
	title VARCHAR(255) NOT NULL,
	description VARCHAR(255),
	completed BOOL NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
ALTER TABLE todos OWNER TO nottodo;

CREATE INDEX IF NOT EXISTS idx_todos_completed ON todos (completed);
CREATE INDEX IF NOT EXISTS idx_todos_created_at ON todos (created_at);

COMMENT ON TABLE todos IS '待办事项';
COMMENT ON COLUMN todos.id IS 'ID';
COMMENT ON COLUMN todos.title IS '标题';
COMMENT ON COLUMN todos.description IS '描述';
COMMENT ON COLUMN todos.completed IS '是否完成';
COMMENT ON COLUMN todos.created_at IS '创建时间';

-- 用户 ===============================
CREATE TABLE IF NOT EXISTS users (
	id BIGSERIAL PRIMARY KEY,
	username VARCHAR(255) UNIQUE NOT NULL,
	password VARCHAR(255) NOT NULL,
    nickname VARCHAR(255),
    register_at TIMESTAMPTZ,
    is_validated BOOL NOT NULL DEFAULT FALSE,
    validated_at TIMESTAMPTZ,
	created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
ALTER TABLE users OWNER TO nottodo;

CREATE INDEX IF NOT EXISTS idx_users_username ON users (username);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users (created_at);

COMMENT ON TABLE users IS '用户';
COMMENT ON COLUMN users.id IS 'ID';
COMMENT ON COLUMN users.username IS '用户名';
COMMENT ON COLUMN users.password IS '密码';
COMMENT ON COLUMN users.nickname IS '昵称';
COMMENT ON COLUMN users.register_at IS '注册时间';
COMMENT ON COLUMN users.is_validated IS '是否已通过验证';
COMMENT ON COLUMN users.validated_at IS '验证时间';
COMMENT ON COLUMN users.created_at IS '创建时间';

-- 邮箱 ===============================
CREATE TABLE IF NOT EXISTS emails (
	id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    is_primary BOOL NOT NULL DEFAULT FALSE,
    is_verified BOOL NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
ALTER TABLE emails OWNER TO nottodo;

CREATE INDEX IF NOT EXISTS idx_emails_user_id ON emails (user_id);
CREATE INDEX IF NOT EXISTS idx_emails_email ON emails (email);

COMMENT ON TABLE emails IS '邮箱';
COMMENT ON COLUMN emails.id IS 'ID';
COMMENT ON COLUMN emails.user_id IS '用户ID';
COMMENT ON COLUMN emails.email IS '邮箱地址';
COMMENT ON COLUMN emails.is_primary IS '是否为主要邮箱';
COMMENT ON COLUMN emails.is_verified IS '是否已验证';
COMMENT ON COLUMN emails.created_at IS '创建时间';
