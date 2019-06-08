package utils

import (
	"strconv"

	cache "github.com/patrickmn/go-cache"
	"github.com/sevings/mindwell-server/models"
)

func loadImageNotCached(srv *MindwellServer, tx *AutoTx, imageID int64) *models.Image {
	baseURL := srv.ConfigString("images.base_url")

	var authorID int64
	var path, extension string

	tx.Query("SELECT user_id, path, extension FROM images WHERE id = $1", imageID).Scan(&authorID, &path, &extension)
	if authorID == 0 {
		return nil
	}

	img := &models.Image{
		ID: imageID,
		Author: &models.User{
			ID: authorID,
		},
		Type: extension,
	}

	var width, height int64
	var size string
	tx.Query(`
		SELECT width, height, (SELECT type FROM size WHERE size.id = image_sizes.size)
		FROM image_sizes
		WHERE image_sizes.image_id = $1
	`, imageID)

	for tx.Scan(&width, &height, &size) {
		switch size {
		case "thumbnail":
			img.Thumbnail = &models.ImageSize{
				Height: height,
				Width:  width,
				URL:    baseURL + "albums/thumbnails/" + path,
			}
		case "small":
			img.Small = &models.ImageSize{
				Height: height,
				Width:  width,
				URL:    baseURL + "albums/small/" + path,
			}
		case "medium":
			img.Medium = &models.ImageSize{
				Height: height,
				Width:  width,
				URL:    baseURL + "albums/medium/" + path,
			}
		case "large":
			img.Large = &models.ImageSize{
				Height: height,
				Width:  width,
				URL:    baseURL + "albums/large/" + path,
			}
		}
	}

	return img
}

func LoadImage(srv *MindwellServer, tx *AutoTx, imageID int64) *models.Image {
	idStr := strconv.FormatInt(imageID, 10)
	oldImg, found := srv.Imgs.Get(idStr)
	if found {
		return oldImg.(*models.Image)
	}

	img := loadImageNotCached(srv, tx, imageID)
	if img == nil {
		return img
	}

	srv.Imgs.Set(idStr, img, cache.DefaultExpiration)
	return img
}
