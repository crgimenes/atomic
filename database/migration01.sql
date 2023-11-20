CREATE TABLE IF NOT EXISTS %s_users (
    id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
	nickname TEXT NOT NULL UNIQUE,
	email TEXT NOT NULL UNIQUE,
	password TEXT, -- can be empty if using ssh key
	ssh_public_key TEXT, -- can be empty if using password
	groups TEXT NOT NULL DEFAULT 'users', -- users,sysop
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

