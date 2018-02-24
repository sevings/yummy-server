package test

import (
	"testing"

	"github.com/sevings/yummy-server/models"

	"github.com/sevings/yummy-server/restapi/operations/design"
	"github.com/stretchr/testify/require"
)

func TestDesign(t *testing.T) {
	req := require.New(t)
	userDesign := profiles[0].Design

	{
		load := api.DesignGetDesignHandler.Handle
		resp := load(design.GetDesignParams{}, userIDs[0])
		body, ok := resp.(*design.GetDesignOK)
		req.True(ok)
		req.Equal(userDesign, body.Payload)
	}

	{
		backColor := "#373737"
		textColor := "#c8c8c8"

		userDesign.BackgroundColor = models.Color(backColor)
		userDesign.CSS = "a { color : gray;	}"
		userDesign.FontFamily = "Arial" //! \todo other fonts
		userDesign.FontSize = 50
		userDesign.TextAlignment = models.DesignTextAlignmentCenter
		userDesign.TextColor = models.Color(textColor)

		params := design.PutDesignParams{
			BackgroundColor: &backColor,
			CSS:             &userDesign.CSS,
			FontFamily:      &userDesign.FontFamily,
			FontSize:        &userDesign.FontSize,
			TextAlignment:   &userDesign.TextAlignment,
			TextColor:       &textColor,
		}

		edit := api.DesignPutDesignHandler.Handle
		resp := edit(params, userIDs[0])
		body, ok := resp.(*design.PutDesignOK)
		req.True(ok)
		req.Equal(userDesign, body.Payload)
	}

	{
		backColor := "#aaaaaa"
		userDesign.BackgroundColor = models.Color(backColor)
		userDesign.CSS = ""

		params := design.PutDesignParams{
			BackgroundColor: &backColor,
			CSS:             &userDesign.CSS,
		}

		edit := api.DesignPutDesignHandler.Handle
		resp := edit(params, userIDs[0])
		body, ok := resp.(*design.PutDesignOK)
		req.True(ok)
		req.Equal(userDesign, body.Payload)
	}
}
