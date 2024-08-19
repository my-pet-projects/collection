package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"
	"os"
	"path/filepath"

	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/my-pet-projects/collection/internal/config"
	"github.com/my-pet-projects/collection/internal/db"
	"github.com/my-pet-projects/collection/internal/log"
	"github.com/my-pet-projects/collection/internal/model"
	"github.com/my-pet-projects/collection/internal/service"
	"github.com/my-pet-projects/collection/internal/storage"
)

func main() { //nolint:funlen
	ctx := context.Background()
	cfg, cfgErr := config.NewConfig()
	if cfgErr != nil {
		panic(cfgErr)
	}
	logger := log.NewLogger(cfg)
	dbClient, dbClientErr := db.NewClient(cfg)
	if dbClientErr != nil {
		panic(dbClientErr)
	}
	mediaStore := db.NewMediaStore(dbClient, logger)
	beerMediaStore := db.NewBeerMediaStore(dbClient, logger)
	sdkConfig, sdkConfigErr := awscfg.LoadDefaultConfig(ctx,
		awscfg.WithRegion(cfg.AwsConfig.Region),
		awscfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AwsConfig.AccessKey, cfg.AwsConfig.SecretKey, "")),
	)
	if sdkConfigErr != nil {
		panic(sdkConfigErr)
	}
	s3Client := s3.NewFromConfig(sdkConfig)
	s3Storage := storage.NewS3Storage(s3Client, logger)
	imageService := service.NewImageService(&mediaStore, &beerMediaStore, &s3Storage, logger)
	fmt.Println(imageService)

	unknowns := 0
	current := 0
	count := 0
	dir := os.DirFS("/Users/I530914/Pictures/all_imgs")
	fs.WalkDir(dir, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		current++

		// if _, err := os.Stat(filepath.Join("/Users/I530914/Pictures/all_imgs", path)); errors.Is(err, os.ErrNotExist) {
		// 	if _, err := os.Stat(filepath.Join("/Users/I530914/Pictures/all_imgs", strings.ReplaceAll(path, "jpg", "png"))); errors.Is(err, os.ErrNotExist) {
		// 		if _, err := os.Stat(filepath.Join("/Users/I530914/Pictures/all_imgs", strings.ReplaceAll(path, "jpeg", "png"))); errors.Is(err, os.ErrNotExist) {
		// 			fmt.Println(path)
		// 			data, _ := ioutil.ReadFile(filepath.Join("/Users/I530914/Pictures/all_imgs_bkp", path))
		// 			_ = ioutil.WriteFile(filepath.Join("/Users/I530914/Pictures/all_imgs_tofix", path), data, 0644)
		// 		}
		// 	}
		// }
		// return nil

		b, readErr := fs.ReadFile(dir, path)
		if err != nil {
			panic(readErr)
		}

		image, _, decodeErr := image.Decode(bytes.NewReader(b))
		if decodeErr != nil {
			panic(decodeErr)
		}
		imageMetadata := model.MediaMetadata{
			Width:  image.Bounds().Dx(),
			Height: image.Bounds().Dy(),
		}

		contentType := ""
		if filepath.Ext(path) == ".png" {
			contentType = "image/png"
		}
		if contentType == "" {
			panic("unknown content type")
		}

		fmt.Println(path)
		fmt.Println(imageMetadata)
		if imageMetadata.Width == imageMetadata.Height && imageMetadata.Height == 800 {
			fmt.Println("Cap")
		} else if imageMetadata.Width == 1000 || imageMetadata.Width == 1500 || imageMetadata.Width == 2000 {
			fmt.Println("Label")
			// if item.Metadata.Width == 1500 {
			// 	count++
			// 	err := os.Rename(filepath.Join("/Users/I530914/Pictures/all_imgs", path), filepath.Join("/Users/I530914/Pictures/all_imgs_tofix", path))
			// 	if err != nil {
			// 	}
			// }
		} else if imageMetadata.Width == 138 && imageMetadata.Height == 400 {
			fmt.Println("Bottle")
		} else {
			fmt.Println(path)
			panic("Unknown")
		}

		items := []model.UploadFormValues{
			{
				Filename:    path,
				Content:     b,
				ContentType: contentType,
			},
		}
		if 1 == 0 {
			uploadErr := imageService.UploadImage(ctx, items)
			if uploadErr != nil {
				panic(uploadErr)
			}
		}

		fmt.Printf("------  %d  ------\n", current)
		return nil
	})

	fmt.Println("count:", count)
	fmt.Println("Unknowns:", unknowns)
}
