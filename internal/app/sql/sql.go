package sql

const (
	// =================== TASKS ===================

	CreateTasks = `
		CREATE TABLE IF NOT EXISTS tasks
		(
			id UINT UNIQUE PRIMARY KEY,
			name TEXT,
			group_id INT,
			story_points INT,
			priority UINT,
			description TEXT,
			connections TEXT,
			reward UINT,
			type TEXT,
			regularity TEXT,
			expirience TEXT
		);
	`

	SelectAllTasks = `
		SELECT * FROM tasks;
	`

	SelectTaskById = `
		SELECT *
		FROM tasks
		WHERE id = $1;
	`

	SelectTasksByGroup = `
		SELECT *
		FROM tasks
		WHERE group = $1;
	`

	// =================== GROUPS ===================

	CreateGroups = `
		CREATE TABLE IF NOT EXISTS groups
		(
			id UINT,
			name TEXT,
			color TEXT
			level UINT,
			exp_need UINT,
			exp_have UINT
		);
	`

	SelectAllGroups = `
		SELECT * FROM groups;
	`
)
