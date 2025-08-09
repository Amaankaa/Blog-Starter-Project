package infrastructure

import (
    "context"
    "fmt"
    "mime/multipart"

    "github.com/cloudinary/cloudinary-go/v2"
    "github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type ICloudinaryService interface {
    UploadImage(ctx context.Context, file multipart.File, filename string) (string, error)
}

type CloudinaryService struct {
    cld *cloudinary.Cloudinary
}

func NewCloudinaryService(cloudName, apiKey, apiSecret string) *CloudinaryService {
    cld, _ := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
    return &CloudinaryService{cld: cld}
}

func (cs *CloudinaryService) UploadImage(ctx context.Context, file multipart.File, filename string) (string, error) {
    uploadParams := uploader.UploadParams{
        PublicID: fmt.Sprintf("profile_pictures/%s", filename),
        Folder:   "blog_app/profiles",
    }

    result, err := cs.cld.Upload.Upload(ctx, file, uploadParams)
    if err != nil {
        return "", err
    }

    return result.SecureURL, nil
}