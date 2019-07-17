package users

import (
	"fmt"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/users"
	"github.com/sevings/mindwell-server/utils"
)

func searchUsers(srv *utils.MindwellServer, tx *utils.AutoTx, params users.GetUsersParams) middleware.Responder {
	const query = usersQuerySelect + `
					FROM (
						SELECT *, $1 <<-> to_search_string(name, show_name, country, city) AS trgm_dist 
						FROM users 
						ORDER BY trgm_dist
						LIMIT 50
					) AS users, gender, user_privacy
					WHERE trgm_dist < 0.6` + usersQueryJoins
	tx.Query(query, params.Query)
	list := loadUserList(srv, tx)
	body := &users.GetUsersOKBody{
		Users: list,
		Query: *params.Query,
	}
	return users.NewGetUsersOK().WithPayload(body)

}

func loadTopUsers(srv *utils.MindwellServer, tx *utils.AutoTx, params users.GetUsersParams) middleware.Responder {
	query := usersQuerySelect

	if *params.Top == "rank" {
		query += "FROM users, gender, user_privacy WHERE invited_by IS NOT NULL" + usersQueryJoins + "ORDER BY rank ASC"
	} else if *params.Top == "new" {
		query += "FROM users, gender, user_privacy WHERE invited_by IS NOT NULL" + usersQueryJoins + " ORDER BY created_at DESC"
	} else if *params.Top == "waiting" {
		query += `
			FROM (
				SELECT *, COALESCE(last_entries.last_created_at, '-infinity') AS last_entry
				FROM users
				LEFT JOIN (
					SELECT author_id, MAX(entries.created_at) AS last_created_at
					FROM entries 
					INNER JOIN users ON entries.author_id = users.id
					INNER JOIN entry_privacy ON entries.visible_for = entry_privacy.id
					INNER JOIN user_privacy ON users.privacy = user_privacy.id
					WHERE entry_privacy.type = 'all' 
						AND (user_privacy.type = 'all' OR user_privacy.type = 'invited')
					GROUP BY author_id
				) AS last_entries ON users.id = last_entries.author_id
				WHERE invited_by IS NULL
				ORDER BY last_entry DESC, created_at DESC
			) as users
			INNER JOIN gender ON gender.id = users.gender
			INNER JOIN user_privacy ON users.privacy = user_privacy.id
		`
	} else {
		fmt.Printf("Unknown users top: %s\n", *params.Top)
		return users.NewGetUsersOK() //.WithPayload(srv.NewError(nil))
	}

	query += " LIMIT 50"
	tx.Query(query)
	list := loadUserList(srv, tx)
	body := &users.GetUsersOKBody{
		Users: list,
		Top:   *params.Top,
	}
	return users.NewGetUsersOK().WithPayload(body)
}

func newUsersLoader(srv *utils.MindwellServer) func(users.GetUsersParams, *models.UserID) middleware.Responder {
	return func(params users.GetUsersParams, userID *models.UserID) middleware.Responder {
		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			if params.Query != nil {
				return searchUsers(srv, tx, params)
			} else {
				return loadTopUsers(srv, tx, params)
			}
		})
	}
}
