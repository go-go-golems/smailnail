package enrich

// SchemaMigrationV2Statements returns the schema additions used by the
// enrichment passes. Mirror bootstrapping owns applying them.
func SchemaMigrationV2Statements() []string {
	return []string{
		`ALTER TABLE messages ADD COLUMN thread_id TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE messages ADD COLUMN thread_depth INTEGER NOT NULL DEFAULT 0`,
		`CREATE INDEX IF NOT EXISTS idx_messages_thread_id ON messages(thread_id)`,
		`CREATE TABLE IF NOT EXISTS threads (
			thread_id TEXT PRIMARY KEY,
			subject TEXT NOT NULL DEFAULT '',
			account_key TEXT NOT NULL DEFAULT '',
			mailbox_name TEXT NOT NULL DEFAULT '',
			message_count INTEGER NOT NULL DEFAULT 0,
			participant_count INTEGER NOT NULL DEFAULT 0,
			first_sent_date TEXT NOT NULL DEFAULT '',
			last_sent_date TEXT NOT NULL DEFAULT '',
			last_rebuilt_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS senders (
			email TEXT PRIMARY KEY,
			display_name TEXT NOT NULL DEFAULT '',
			domain TEXT NOT NULL DEFAULT '',
			is_private_relay BOOLEAN NOT NULL DEFAULT FALSE,
			relay_display_domain TEXT NOT NULL DEFAULT '',
			msg_count INTEGER NOT NULL DEFAULT 0,
			first_seen_date TEXT NOT NULL DEFAULT '',
			last_seen_date TEXT NOT NULL DEFAULT '',
			unsubscribe_mailto TEXT NOT NULL DEFAULT '',
			unsubscribe_http TEXT NOT NULL DEFAULT '',
			has_list_unsubscribe BOOLEAN NOT NULL DEFAULT FALSE,
			last_synced_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_senders_domain ON senders(domain)`,
		`CREATE INDEX IF NOT EXISTS idx_senders_is_relay ON senders(is_private_relay)`,
		`ALTER TABLE messages ADD COLUMN sender_email TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE messages ADD COLUMN sender_domain TEXT NOT NULL DEFAULT ''`,
	}
}
