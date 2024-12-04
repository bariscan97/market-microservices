package utils

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"strings"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func UploadToCloudinary(file *multipart.FileHeader) (string, error) {
	
	defer func() {
		os.RemoveAll("uploads/")
	}()

	if strings.HasSuffix(file.Filename, ".png")  &&  strings.HasSuffix(file.Filename, ".jpg")  {
		return "", fmt.Errorf("unexpected file")
	}
	
	cloudinary_url := os.Getenv("CLOUDINARY_URL")

	cld, _ := cloudinary.NewFromURL(cloudinary_url)

	ctx := context.Background()

	resp, err := cld.Upload.Upload(ctx,
		"uploads/"+file.Filename,
		uploader.UploadParams{},
	)

	if err != nil {
        log.Fatal(err)
		return "", fmt.Errorf("some error:%s", err)
	}

	return resp.URL, nil
}