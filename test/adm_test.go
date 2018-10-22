package test

import (
	"testing"

	"github.com/sevings/mindwell-server/models"

	"github.com/sevings/mindwell-server/restapi/operations/adm"
	"github.com/stretchr/testify/require"
)

func checkGrandsonAddress(t *testing.T, userID *models.UserID, name, postcode, country, address string, anonymous bool) {
	req := require.New(t)

	load := api.AdmGetAdmGrandsonHandler.Handle
	resp := load(adm.GetAdmGrandsonParams{}, userID)
	body, ok := resp.(*adm.GetAdmGrandsonOK)
	req.True(ok)

	addr := body.Payload
	req.Equal(name, addr.Name)
	req.Equal(postcode, addr.Postcode)
	req.Equal(country, addr.Country)
	req.Equal(address, addr.Address)
	req.Equal(anonymous, addr.Anonymous)
}

func updateGrandsonAddress(t *testing.T, userID *models.UserID, name, postcode, country, address string, anonymous bool) {
	params := adm.PostAdmGrandsonParams{
		Name:      name,
		Postcode:  postcode,
		Country:   country,
		Address:   address,
		Anonymous: &anonymous,
	}

	post := api.AdmPostAdmGrandsonHandler.Handle
	resp := post(params, userID)
	_, ok := resp.(*adm.PostAdmGrandsonOK)
	require.True(t, ok)

	checkGrandsonAddress(t, userID, name, postcode, country, address, anonymous)
}

func TestAdm(t *testing.T) {
	checkGrandsonAddress(t, userIDs[0], "", "", "", "", false)
	updateGrandsonAddress(t, userIDs[0], "aaa", "213", "Aaa", "aaaa", false)
	updateGrandsonAddress(t, userIDs[0], "bbb", "5654", "Bbb", "bbbb", true)
}
