package main

import (
	"fmt"
	"log"
	"strconv"

	"example.com/uploader"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// func main() {
// 	mux := http.NewServeMux()
// 	// mux.HandleFunc("/", indexHandler)
// 	mux.HandleFunc("/v1/file", uploadFile2)

// 	if err := http.ListenAndServe(":4500", mux); err != nil {
// 		log.Fatal(err)
// 	}
// }

func main() {
	// create new fiber instance  and use across whole app
	app := fiber.New()

	// middleware to allow all clients to communicate using http and allow cors
	app.Use(cors.New())

	// handle image uploading using put request
	app.Put("/v1/file", uploadFile)

	// handle image uploading using put request
	app.Get("/v1/:url", getFile)

	log.Fatal(app.Listen(":4500"))
}

//const MAX_UPLOAD_SIZE = 1024 * 1024 // 1MB
const DEFAULT_EXPIRATION_TIME int64 = 1 // 1 MINUTE

func uploadFile(c *fiber.Ctx) error {
	// parse incomming image file
	fileHeader, err := c.FormFile("file")
	if err != nil {
		log.Println("image upload error --> ", err)
		return c.JSON(fiber.Map{"status": 500, "message": "Server error", "data": nil})
	}

	fmt.Println("Start upload file with name: ", fileHeader.Filename)
	expirationTime := getExpirationTime(c)
	if expirationTime == -1 {
		return c.JSON(fiber.Map{"status": 500, "message": err, "data": nil})
	}

	file, err := fileHeader.Open()
	if err != nil {
		log.Fatal("Failed to get file", err)
		fmt.Println("Failed to get file", err)

		return c.JSON(fiber.Map{"status": 500, "message": err, "data": nil})
	}

	defer file.Close()

	filepath, err := uploader.Upload(file, fileHeader, expirationTime)
	if err != nil {
		// maybe if s3 is not accessible we will save the file localy till further notice
		// err = c.SaveFile(fileHeader, "./images/{name}.jpg")
		// if err != nil {
		// 	fmt.Println("Save file err", err)
		// }

		log.Println("Failed to upload file", err)
		return c.JSON(fiber.Map{"status": 500, "message": err, "data": nil})
	}

	return c.JSON(fiber.Map{"status": 201, "message": "Image uploaded successfully", "data": filepath})
}

// Returns the file that was uploaded
func getFile(c *fiber.Ctx) error {
	params := c.AllParams()

	fmt.Println("Name ", params["url"])

	url, err := uploader.GetObject(params["url"])

	if err != nil {
		log.Println("Failed to retrieve file", err)
		return c.SendString("")
	}

	return c.SendString(url)
}

func getExpirationTime(c *fiber.Ctx) int64 {
	expirationTimeHeader := c.GetReqHeaders()["Expiration"]
	fmt.Println("Expiration: ", expirationTimeHeader)

	expirationTime := DEFAULT_EXPIRATION_TIME
	if len(expirationTimeHeader) > 0 {
		expirationTimeInt, err := strconv.ParseInt(expirationTimeHeader, 10, 64)
		if err != nil {
			log.Fatal("expirationTimeInt error --> ", err)
			fmt.Println("expirationTimeInt error: ", err)

			return -1
		}

		log.Println("ExpirationTimeHeader is set to ", expirationTimeHeader)

		fmt.Println("ExpirationTimeHeader is set to ", expirationTimeHeader)
		expirationTime = expirationTimeInt
	}
	return expirationTime
}
