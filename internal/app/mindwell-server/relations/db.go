package relations

import (
	"database/sql"
	"strings"

	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/utils"
)

func relationship(tx *utils.AutoTx, from, to string) *models.Relationship {
	const q = `
		SELECT relation.type
		FROM relation, relations
		WHERE from_id = (SELECT id FROM users where lower(name) = lower($1)) 
			AND to_id = (SELECT id FROM users where lower(name) = lower($2)) 
			AND relations.type = relation.id`

	relation := models.Relationship{
		From: from,
		To:   to,
	}

	tx.Query(q, from, to).Scan(&relation.Relation)
	if tx.Error() == sql.ErrNoRows {
		relation.Relation = models.RelationshipRelationNone
	}

	return &relation
}

func setRelationship(tx *utils.AutoTx, from, to, relation string) (*models.Relationship, bool) {
	const q = `
		INSERT INTO relations (from_id, to_id, type)
		VALUES ((SELECT id FROM users where lower(name) = lower($1)), 
				(SELECT id FROM users where lower(name) = lower($2)), 
				(SELECT id FROM relation WHERE type = $3))
		ON CONFLICT ON CONSTRAINT unique_relation
		DO UPDATE SET type = EXCLUDED.type, changed_at = CURRENT_TIMESTAMP`

	tx.Exec(q, from, to, relation)

	return &models.Relationship{
		From:     from,
		To:       to,
		Relation: relation,
	}, tx.RowsAffected() == 1
}

func removeRelationship(tx *utils.AutoTx, from, to string) *models.Relationship {
	const q = `
		DELETE FROM relations
		WHERE from_id = (SELECT id FROM users where lower(name) = lower($1)) 
			AND to_id = (SELECT id FROM users where lower(name) = lower($2))`

	tx.Exec(q, from, to)

	return &models.Relationship{
		From:     from,
		To:       to,
		Relation: models.RelationshipRelationNone,
	}
}

func isPrivateTlog(tx *utils.AutoTx, name string) bool {
	const q = `
		SELECT user_privacy.type = 'followers'
		FROM users, user_privacy
		WHERE lower(users.name) = lower($1) AND users.privacy = user_privacy.id`

	var private bool
	tx.Query(q, name).Scan(&private)

	return private
}

func removeInvite(tx *utils.AutoTx, invite string, userID int64) bool {
	words := strings.Fields(invite)
	if len(words) != 3 {
		return false
	}

	const q = `
		DELETE FROM invites
		WHERE referrer_id = $1
		  	AND word1 = (SELECT id FROM invite_words WHERE word = $2)
		  	AND word2 = (SELECT id FROM invite_words WHERE word = $3)
			AND word3 = (SELECT id FROM invite_words WHERE word = $4)
		`

	tx.Exec(q, userID,
		strings.ToLower(words[0]),
		strings.ToLower(words[1]),
		strings.ToLower(words[2]))

	return tx.RowsAffected() == 1
}

func isTlogExistsAndInvited(tx *utils.AutoTx, name string) (bool, bool) {
	var exists, invited bool
	tx.Query("SELECT true, invited_by IS NOT NULL FROM users WHERE lower(name) = lower($1)", name).Scan(&exists, &invited)
	return exists, invited
}

func canInvite(tx *utils.AutoTx, name string) bool {
	q := `
		WITH tlog AS (
			SELECT id, up_votes, down_votes
			FROM entries
			WHERE author_id = (SELECT id FROM users WHERE lower(name) = lower($1))
		),
		sumVotes AS (
			SELECT sum(up_votes - down_votes) >= 5 AS ok
			FROM tlog
		),
		countVotes AS (
			SELECT count(distinct entry_votes.user_id) >= 3 AS ok
			FROM tlog
			INNER JOIN entry_votes ON tlog.id = entry_id
			WHERE vote > 0
		)
		SELECT sumVotes.ok AND countVotes.ok
		FROM sumVotes, countVotes
	`

	var allowed bool
	tx.Query(q, name).Scan(&allowed)
	return allowed
}

func setInvited(tx *utils.AutoTx, from int64, to string) {
	tx.Query("UPDATE users SET invited_by = $1 WHERE name = $2", from, to)
}
