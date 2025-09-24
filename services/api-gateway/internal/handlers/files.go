package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mtk14m/mini-cloud/api-gateway/internal/config"
)

func UploadFile(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Récupérer l'utilisateur depuis le contexte
		userID := c.GetString("user_id")
		username := c.GetString("username")

		// Récupérer le fichier
		file, header, err := c.Request.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
			return
		}
		defer file.Close()

		// TODO: Appeler le service de fichiers
		// Pour l'instant, on simule
		c.JSON(http.StatusOK, gin.H{
			"message":     "File uploaded successfully",
			"filename":    header.Filename,
			"size":        header.Size,
			"uploaded_by": username,
			"user_id":     userID,
		})
	}
}

func DownloadFile(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		fileID := c.Param("id")
		userID := c.GetString("user_id")

		// TODO: Appeler le service de fichiers
		// Pour l'instant, on simule
		c.JSON(http.StatusOK, gin.H{
			"message": "File download initiated",
			"file_id": fileID,
			"user_id": userID,
		})
	}
}

func DeleteFile(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		fileID := c.Param("id")
		userID := c.GetString("user_id")

		// TODO: Appeler le service de fichiers
		// Pour l'instant, on simule
		c.JSON(http.StatusOK, gin.H{
			"message":    "File deleted successfully",
			"file_id":    fileID,
			"deleted_by": userID,
		})
	}
}
