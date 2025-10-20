CREATE TABLE IF NOT EXISTS todos (
	id BIGSERIAL PRIMARY KEY,
	title VARCHAR(255) NOT NULL,
	description VARCHAR(255),
	completed BOOL NOT NULL,
	created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
COMMENT ON COLUMN todos.id IS 'ID';
COMMENT ON COLUMN todos.title IS '标题';
COMMENT ON COLUMN todos.description IS '描述';
COMMENT ON COLUMN todos.completed IS '是否完成';
COMMENT ON COLUMN todos.created_at IS '创建时间';


CREATE TABLE IF NOT EXISTS settings (
    key VARCHAR(255) PRIMARY KEY,
    value TEXT NOT NULL,
    description TEXT,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON COLUMN settings.key IS '键';
COMMENT ON COLUMN settings.value IS '值';
COMMENT ON COLUMN settings.description IS '描述';
COMMENT ON COLUMN settings.updated_at IS '更新时间';
