package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"slices"
	"time"

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

type Image struct {
	ID struct {
		Oid string `json:"$oid"`
	} `json:"_id"`
	Filename  string `json:"filename"`
	Size      int    `json:"size"`
	Hash      string `json:"hash"`
	CreatedAt struct {
		Date time.Time `json:"$date"`
	} `json:"createdAt"`
	Type       int    `json:"type"`
	MimeType   string `json:"mimeType"`
	ModifiedAt struct {
		Date time.Time `json:"$date"`
	} `json:"modifiedAt,omitempty"`
}

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
	beerStore := db.NewBeerStore(dbClient, logger)
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

	beerFilter := model.BeerFilter{}
	pagination, _ := beerStore.PaginateBeers(beerFilter)
	beers := pagination.Results
	fmt.Println(len(beers))

	// jsonBytes, err := os.ReadFile("images.json")
	// if err != nil {
	// 	panic(err)
	// }
	// var jsonImages []Image
	// jsonErr := json.Unmarshal(jsonBytes, &jsonImages)
	// if jsonErr != nil {
	// 	panic(jsonErr)
	// }

	unknowns := 0
	current := 0
	count := 0

	strangeImages := []string{
		"047475c3ca904d8c8dedb5048d7cf1d11c1fd0bc.png",
		"0c57d25d741aa1fc30fa05a20fe53356b5da523b.png",
		"10795e8e16dbcfb604a956cf8776d8e8cdfa36a9.png",
		// "134667b7c53f7471e020998361560a50a2cc4061.png",
		"1c327eb7f0702380e707730056d821ccb7f04158.png",
		"22b471df88223815c0251d077fe603f59d98010b.png",
		"25a4f1873e6d82ecc7b8a059b8d5e119012e0df5.png",
		"274749e378c5444c98d65dca8b90f01d9acb86d5.png",
		"351685461711c519fa85e9bc64b332c8fb326e97.png",
		"35b35acb06019eb8691cccbe1fcc3df98de4d7b0.png",
		"3f2ffd05b60f03a1c5986cef75495414c4ef060d.png",
		"41d864f6cbf9cf032a0d1c2cd8fbd2eb043d1d41.png",
		"4293f4f35555c1c20c24dfac0c15757ae07f17fe.png",
		"474eea88bc0ab49d66d8712e964a0a66e913bb2d.png",
		"487d192c92119ee4a16a4666973ac2ff9c91c379.png",
		"4a08edb6390a1f910ec790d19a82df832f574291.png",
		"4fb7683f802485155a2c61ba9d32cdbfb20e1c1d.png",
		"5b7c40c4213d1809f218c9b91b60c513973ddb19.png",
		"5eea3a0b6e47cf94ce1784982e0c87a19329f9aa.png",
		"63c496e60ee650faf35c4e2c02b08a175203a8e7.png",
		"67d7ab55fedcc19829a431df0cf6ea06921f0176.png",
		"6c295de0819fc269a01aa3565020c817cc4dffc2.png",
		"6cd18cbda2600f79658d115fc40f39b56f88c096.png",
		"6f1b608846e87398cf3ebb8aefe4a065e246014a.png",
		"6f27a78341e01e272567fb6d849bb23a7fe87660.png",
		"7108d32e6544e5ece24676806bdc05f11387f907.png",
		"71b4d8ea1bdd9898a2affa0f02a934f9704ba565.png",
		"7eacdf3512b3527cdafaf45dc0d18ad8c56f12e2.png",
		"807fcff16d5864d198f1c683f6ce5d0633511049.png",
		"859396c9943949b3c1554da3c8848dbd1cb69593.png",
		"8640556112c4c9b6ee3a53b1725dd2b9ca7ff374.png",
		"975e0e0182b2d895d5a53cffc2d751dad470184d.png",
		"a1e80671352a7a1e6ed9ebe8b83756e80d56cc01.png",
		"a3df443c38420b3b069b62992b411baf9bf1be08.png",
		"a913709e291740265991cadf54b0c5f8a77b453d.png",
		"aac49590dfe263aea99fb93039528e5931bf91c0.png",
		"adccf5e583cd35ec2433d5c43348122873582015.png",
		"b0c16db36db858386bfe958b6fe484a7560ba4c5.png",
		"b5a15df17f07bda8b45247a800bea795bbf8d5e7.png",
		"b6c5379d5bf84f6a86bd911c2f6262331d9991db.png",
		"c24f21925a75b4b7c6d86abcf9720de5f34f6578.png",
		"cd435c27759df17046aee5ff4ff2df5a44bbc69e.png",
		"d964c63c865fd3eccb4aa5477c803beb80b9379c.png",
		"e3795c71183550e5d7e79dae123bde21b627a30c.png",
		"ea8cbf98ed02d5df31941a2ab93f04aa8da9a38f.png",
		"ee7949385be83c2aae3110d62291992be205598e.png",
		"f90d8d2845b72cc849d3f94e663197ce0522c2d7.png",
		"fa0eb54b6abb9dacdeeea4df9a1e0fe76d7b6298.png",
		"fa6ff15d3d2b3285000ffca5cb1935e5231b9250.png",
	}

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

		if slices.Contains(strangeImages, path) {
			_ = ioutil.WriteFile(filepath.Join("/Users/I530914/Pictures/strange_imgs", path), b, 0644)
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

		// var jsonImageId string
		// for _, jsonImage := range jsonImages {
		// 	filename := strings.Replace(jsonImage.Filename, ".png", "", 1)
		// 	filename = strings.Replace(filename, ".jpeg", "", 1)
		// 	filename = strings.Replace(filename, ".jpg", "", 1)
		// 	if filename == strings.Replace(path, ".png", "", 1) {
		// 		jsonImageId = jsonImage.ID.Oid
		// 		break
		// 	}
		// }

		// if jsonImageId == "" && !slices.Contains(strangeImages, path) {
		// 	fmt.Printf("Not found: %s, %s\n", path, jsonImageId)
		// 	// strangeImages = append(strangeImages, path)
		// 	panic("Not found")
		// }

		// var beerID int
		// for _, beer := range beers {
		// 	if strings.Contains(*beer.OldImageIds, jsonImageId) {
		// 		beerID = beer.ID
		// 		break
		// 	}
		// }
		// if beerID == 0 {
		// 	fmt.Printf("Beer not found: %s, %s\n", path, jsonImageId)
		// 	// strangeBeers = append(strangeBeers, path)
		// 	panic("Not found")
		// }

		items := []model.UploadFormValues{
			{
				Filename:    path,
				Content:     b,
				ContentType: contentType,
				// BeerID:      &beerID,
			},
		}

		// if current < 2500 {
		uploadErr := imageService.UploadImage(ctx, items)
		if uploadErr != nil {
			panic(uploadErr)
		}
		// }

		fmt.Printf("------  %d  ------\n", current)
		return nil
	})

	fmt.Println("count:", count)
	fmt.Println("Unknowns:", unknowns)

	fmt.Println("Strange images:", strangeImages)
}
