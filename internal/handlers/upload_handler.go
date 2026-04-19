// internal/handlers/upload_handler.go
package handlers

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/shreyas100-hobby/ecommerce-backend/internal/services"
)

type UploadHandler struct {
    cloudinary *services.CloudinaryService
}

func NewUploadHandler(cloudinary *services.CloudinaryService) *UploadHandler {
    return &UploadHandler{cloudinary: cloudinary}
}

// UploadImage handles both file upload and URL upload
func (h *UploadHandler) UploadImage(c *gin.Context) {
    ctx := c.Request.Context()

    // Option A — URL upload
    imageURL := c.PostForm("url")
    if imageURL != "" {
        url, publicID, err := h.cloudinary.UploadFromURL(ctx, imageURL)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "error": "Failed to upload from URL",
            })
            return
        }
        c.JSON(http.StatusOK, gin.H{
            "url":       url,
            "public_id": publicID,
        })
        return
    }

    // Option B — File upload
    file, header, err := c.Request.FormFile("image")
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Please provide an image file or URL",
        })
        return
    }
    defer file.Close()

    // Validate file type
    contentType := header.Header.Get("Content-Type")
    if !strings.HasPrefix(contentType, "image/") {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Only image files are allowed",
        })
        return
    }

    // Max 10MB
    if header.Size > 10*1024*1024 {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Image must be less than 10MB",
        })
        return
    }

    url, publicID, err := h.cloudinary.UploadFile(ctx, file, header.Filename)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to upload image",
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "url":       url,
        "public_id": publicID,
    })
}