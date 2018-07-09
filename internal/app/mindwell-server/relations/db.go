package relations

import (
	"database/sql"

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

func sendNewFollower(srv *utils.MindwellServer, tx *utils.AutoTx, toPrivate bool, from, to string) {
	const toQ = `
		SELECT show_name, email, verified
		FROM users 
		WHERE lower(name) = lower($1)
	`

	var toShowName, email string
	var verified bool
	tx.Query(toQ, to).Scan(&toShowName, &email, &verified)
	if !verified {
		return
	}

	const fromQ = `
		SELECT show_name, gender.type 
		FROM users, gender 
		WHERE lower(users.name) = lower($1) AND users.gender = gender.id`

	var fromShowName, fromGender string
	tx.Query(fromQ, from).Scan(&fromShowName, &fromGender)

	srv.Mail.SendNewFollower(email, from, fromShowName, fromGender, toPrivate, toShowName)
}
