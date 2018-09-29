package users

import (
	"fmt"

	"github.com/go-openapi/runtime/middleware"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations/users"
	"github.com/sevings/mindwell-server/utils"
)

func newTopUsersLoader(srv *utils.MindwellServer) func(users.GetUsersParams, *models.UserID) middleware.Responder {
	return func(params users.GetUsersParams, userID *models.UserID) middleware.Responder {
		query := usersQuerySelect + "FROM long_users "

		if *params.Top == "rank" {
			query += "WHERE rank > 0 ORDER BY rank DESC"
		} else if *params.Top == "new" {
			query += "ORDER BY created_at DESC"
		} else {
			fmt.Printf("Unknown users top: %s\n", *params.Top)
			return users.NewGetUsersOK() //.WithPayload(srv.NewError(nil))
		}

		query += " LIMIT 50"

		return srv.Transact(func(tx *utils.AutoTx) middleware.Responder {
			tx.Query(query)
			list := loadUserList(srv, tx)
			body := &users.GetUsersOKBody{
				Users: list,
				Top:   *params.Top,
			}
			return users.NewGetUsersOK().WithPayload(body)
		})
	}
}
