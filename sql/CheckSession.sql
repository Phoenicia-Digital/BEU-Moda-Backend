SELECT id FROM sessions WHERE session_id = $1 AND user_uid = $2 AND expires_at > NOW();