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
						FROM long_users 
						ORDER BY trgm_dist
						LIMIT 50
					) AS long_users
					WHERE trgm_dist < 0.6
					`
	tx.Query(query, params.Query)
	list := loadUserList(srv, tx)
	body := &users.GetUsersOKBody{
		Users: list,
		Query: *params.Query,
	}
	return users.NewGetUsersOK().WithPayload(body)

}

func loadTopUsers(srv *utils.MindwellServer, tx *utils.AutoTx, params users.GetUsersParams) middleware.Responder {
	query := usersQuerySelect + "FROM long_users "

	if *params.Top == "rank" {
		query += "WHERE invited_by IS NOT NULL ORDER BY rank ASC"
	} else if *params.Top == "new" {
		query += "WHERE invited_by IS NOT NULL ORDER BY created_at DESC"
	} else if *params.Top == "waiting" {
		query += "WHERE invited_by IS NULL ORDER BY rank ASC"
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
