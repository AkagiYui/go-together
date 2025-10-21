-- 缓存表，使用 UNLOGGED 关键字创建表
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
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

COMMENT ON TABLE todos IS '待办事项';
COMMENT ON COLUMN todos.id IS 'ID';
COMMENT ON COLUMN todos.title IS '标题';
COMMENT ON COLUMN todos.description IS '描述';
COMMENT ON COLUMN todos.completed IS '是否完成';
COMMENT ON COLUMN todos.created_at IS '创建时间';