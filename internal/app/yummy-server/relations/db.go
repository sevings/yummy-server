package relations

import (
	"database/sql"

	"github.com/sevings/yummy-server/internal/app/yummy-server/utils"
	"github.com/sevings/yummy-server/restapi/operations/relations"
	"github.com/sevings/yummy-server/models"
)

func relationship(tx *utils.AutoTx, from, to int64) *models.Relationship {
	const q = `
		SELECT relation.type
		FROM relation, relations
		WHERE from_id = $1 AND to_id = $2 
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

func setRelationship(tx *utils.AutoTx, from, to int64, relation string) *models.Relationship, bool {
	const q = `
		INSERT INTO relations (from, to, type)
		VALUES ($1, $2, (SELECT id FROM relation WHERE type = $3))
		ON CONFLICT ON CONSTRAINT unique_relation
		DO UPDATE SET type = EXCLUDED.type, changed_at = CURRENT_TIMESTAMP`

	tx.Exec(q, from, to, relation)

	return &models.Relationship{
		From:     from,
		To:       to,
		Relation: relation,
	}, tx.RowsAffected() == 1
}


func removeRelationship(tx *utils.AutoTx, from, to int64) *models.Relationship {
	const q = `
		DELETE FROM relations
		WHERE from = $1 AND to = $2`

	tx.Exec(q, from, to)

	return &models.Relationship{
		From: from,
		To:   to,
		Relation: models.RelationshipRelationNone,
	}
}

func isPrivateTlog(tx *utils.AutoTx, id int64) bool {
	const q = `
		SELECT user_privacy.type = 'followers'
		FROM users, user_privacy
		WHERE users.id = $1 AND users.privacy = user_privacy.id`

	var private bool
	tx.Query(q, id).Scan(&private)

	return private
}
