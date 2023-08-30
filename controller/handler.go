package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"cloud.google.com/go/firestore"
	"github.com/labstack/echo/v4"
	"github.com/nfnt/resize"

	"image"

	_ "image/jpeg"
	_ "image/png"
)

func getInstances(c echo.Context) error {
	return c.JSON(http.StatusOK, instances)
}

func skipRender(c echo.Context) error {
	skip = true
	return c.JSON(http.StatusOK, map[string]interface{}{"status": "ok", "skip": true})
}

func unskipRender(c echo.Context) error {
	skip = false
	return c.JSON(http.StatusOK, map[string]interface{}{"status": "ok", "skip": false})
}

func directRender(client *firestore.Client,
	ledCollection string,
	instanceCollection string,
	mapping []int) func(c echo.Context) error {
	return func(c echo.Context) error {
		skip = true
		pixels := &PixelGrid{}
		// err := c.Bind(pixels)
		// if err != nil {
		// 	log.Println(err)
		// 	return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Invalid request body"}
		// }

		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			log.Println(err)
			return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Invalid request body"}
		}

		err = json.Unmarshal(body, pixels)
		if err != nil {
			log.Println(err)
			return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Invalid request body"}
		}

		ctx := context.Background()
		ledData := &LedData{
			Data: processPixelsForLed(*pixels, mapping),
		}
		_, err = client.Collection(ledCollection).Doc("data").Set(ctx, ledData)
		if err != nil {
			return &echo.HTTPError{Code: http.StatusInternalServerError, Message: "Failed to set led data"}
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}
}

func fileUpload(c echo.Context) error {
	mf, err := c.FormFile("file")
	if err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Invalid request body"}
	}
	file, err := mf.Open()
	if err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Invalid request body"}
	}
	i, _, err := image.Decode(file)
	if err != nil {
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Invalid request body"}
	}
	squareImage := ConvertToSquare(i)
	resized := resize.Resize(127, 127, squareImage, resize.NearestNeighbor)

	pixels := make([][]int, 0)

	gamma := 3.5
	gammaVal := os.Getenv("GAMMA")
	if gammaVal != "" {
		g, err := strconv.ParseFloat(gammaVal, 64)
		if err == nil {
			gamma = g
		}
	}

	log.Println("Gamma:", gamma)

	for j := 0; j < resized.Bounds().Dy(); j += 2 {
		var row []int
		for i := 0; i < resized.Bounds().Dx(); i += 2 {
			r, g, b, _ := resized.At(i, j).RGBA()
			red := gammaCorrect(int(r>>8), gamma)
			green := gammaCorrect(int(g>>8), gamma)
			blue := gammaCorrect(int(b>>8), gamma)
			row = append(row, red<<16+green<<8+blue)
		}
		pixels = append(pixels, row)
	}

	return c.JSON(http.StatusOK, PixelGrid{Pixels: pixels})
}

func gameOfLife(c echo.Context) error {
	pixels := &PixelGrid{}
	// err := c.Bind(pixels)
	// if err != nil {
	// 	log.Println(err)
	// 	return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Invalid request body"}
	// }

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		log.Println(err)
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Invalid request body"}
	}

	err = json.Unmarshal(body, pixels)
	if err != nil {
		log.Println(err)
		return &echo.HTTPError{Code: http.StatusBadRequest, Message: "Invalid request body"}
	}

	nextState := nextGeneration(pixels.Pixels)
	return c.JSON(http.StatusOK, PixelGrid{Pixels: nextState})
}
