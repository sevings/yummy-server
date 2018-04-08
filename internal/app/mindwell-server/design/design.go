package design

import (
	"database/sql"

	"github.com/sevings/mindwell-server/restapi/operations/design"

	"github.com/go-openapi/runtime/middleware"

	"github.com/sevings/mindwell-server/internal/app/mindwell-server/utils"
	"github.com/sevings/mindwell-server/models"
	"github.com/sevings/mindwell-server/restapi/operations"
)

// ConfigureAPI creates operations handlers
func ConfigureAPI(db *sql.DB, api *operations.MindwellAPI) {
	api.DesignGetDesignHandler = design.GetDesignHandlerFunc(newDesignGetter(db))
	api.DesignPutDesignHandler = design.PutDesignHandlerFunc(newDesignEditor(db))
}

func loadDesign(tx *utils.AutoTx, id int64) models.Design {
	const q = `
		SELECT css, 
			background_color, text_color,
			alignment.type, font_size, 
			font_family.type
		FROM users, font_family, alignment
		WHERE users.id = $1 
			AND users.font_family = font_family.id
			AND users.text_alignment = alignment.id`

	var design models.Design
	tx.Query(q, id).Scan(&design.CSS,
		&design.BackgroundColor, &design.TextColor,
		&design.TextAlignment, &design.FontSize,
		&design.FontFamily)

	return design
}

func newDesignGetter(db *sql.DB) func(design.GetDesignParams, *models.UserID) middleware.Responder {
	return func(params design.GetDesignParams, uID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			id := int64(*uID)
			des := loadDesign(tx, id)
			return design.NewGetDesignOK().WithPayload(&des)
		})
	}
}

func editDesign(tx *utils.AutoTx, params design.PutDesignParams, id int64) models.Design {
	design := loadDesign(tx, id)

	if params.CSS != nil {
		design.CSS = *params.CSS
	}

	var backColor string
	if params.BackgroundColor != nil {
		backColor = *params.BackgroundColor
		design.BackgroundColor = models.Color(backColor)
	} else {
		backColor = string(design.BackgroundColor)
	}

	var textColor string
	if params.TextColor != nil {
		textColor = *params.TextColor
		design.TextColor = models.Color(textColor)
	} else {
		textColor = string(design.TextColor)
	}

	design.TextAlignment = params.TextAlignment

	if params.FontFamily != nil {
		design.FontFamily = *params.FontFamily
	}

	if params.FontSize != nil {
		design.FontSize = *params.FontSize
	}

	const q = `
		UPDATE users
		SET css = $2, 
			background_color = $3, text_color = $4,
			text_alignment = (SELECT id FROM alignment WHERE type = $5),
			font_family = (SELECT id FROM font_family WHERE type = $6),
			font_size = $7
		WHERE id = $1`

	tx.Exec(q, id, design.CSS,
		backColor, textColor,
		design.TextAlignment, design.FontFamily, design.FontSize)

	return design
}

func newDesignEditor(db *sql.DB) func(design.PutDesignParams, *models.UserID) middleware.Responder {
	return func(params design.PutDesignParams, uID *models.UserID) middleware.Responder {
		return utils.Transact(db, func(tx *utils.AutoTx) middleware.Responder {
			id := int64(*uID)
			des := editDesign(tx, params, id)
			return design.NewPutDesignOK().WithPayload(&des)
		})
	}
}
