package services

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryService struct {
	cld *cloudinary.Cloudinary
}

func NewCloudinaryService(cloudName, apiKey, apiSecret string) (*CloudinaryService, error) {
	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to init cloudinary: %w", err)
	}
	return &CloudinaryService{cld: cld}, nil
}

func (s *CloudinaryService) UploadFile(
	ctx context.Context,
	file multipart.File,
	fileName string,
) (url string, publicID string, err error) {
	resp, err := s.cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder:         "ecommerce",
		PublicID:       fileName,
		Transformation: "q_auto,f_auto,w_1200",
	})
	if err != nil {
		return "", "", fmt.Errorf("upload failed: %w", err)
	}
	return resp.SecureURL, resp.PublicID, nil
}

func (s *CloudinaryService) UploadFromURL(
	ctx context.Context,
	imageURL string,
) (url string, publicID string, err error) {
	resp, err := s.cld.Upload.Upload(ctx, imageURL, uploader.UploadParams{
		Folder:         "ecommerce",
		Transformation: "q_auto,f_auto,w_1200",
	})
	if err != nil {
		return "", "", fmt.Errorf("upload from url failed: %w", err)
	}
	return resp.SecureURL, resp.PublicID, nil
}

func (s *CloudinaryService) DeleteFile(ctx context.Context, publicID string) error {
	_, err := s.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})
	return err
}