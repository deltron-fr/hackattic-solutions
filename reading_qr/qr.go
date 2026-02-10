package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	_ "image/png"
	"io"
	"net/http"
	"os"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
)

const filePath = "qr.png"

type Response struct {
	ImageURL string `json:"image_url"`
}

type Solution struct {
	Code string `json:"code"`
}

func main() {
	res, err := http.Get("https://hackattic.com/challenges/reading_qr/problem?access_token=8f3a9cb1bc4572d2")
	if err != nil {
		fmt.Printf("error initiating GET request: %v", err)
		return
	}
	defer res.Body.Close()

	var img Response
	err = json.NewDecoder(res.Body).Decode(&img)
	if err != nil {
		fmt.Printf("error decoding json: %v", err)
		return
	}

	err = downloadImage(img.ImageURL)
	if err != nil {
		fmt.Printf("error downloading image: %v", err)
		return
	}

	out, err := getQRCode(filePath)
	if err != nil {
		fmt.Printf("error getting qr code details: %v", err)
		return
	}

	code := Solution{
		Code: out.GetText(),
	}

	buffer := bytes.NewBuffer(nil)
	err = json.NewEncoder(buffer).Encode(code)
	if err != nil {
		fmt.Printf("error encoding json: %v", err)
		return
	}


	req, err := http.NewRequest("POST", "https://hackattic.com/challenges/reading_qr/solve?access_token=8f3a9cb1bc4572d2", buffer)
	if err != nil {
		fmt.Printf("error creating new request: %v", err)
		return
	}

	client := &http.Client{}
	res, err = client.Do(req)
	if err != nil {
		fmt.Printf("error initiating request-response: %v", err)
		return
	}
	defer res.Body.Close()

	var result any
	err = json.NewDecoder(res.Body).Decode(&result)
	if err != nil {
		fmt.Printf("error decoding json: %v", err)
		return
	}

	fmt.Println(result)

}

func downloadImage(imageURL string) error {
	fmt.Println(imageURL)
	res, err := http.Get(imageURL)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, res.Body)
	if err != nil {
		return err
	}

	return nil
}

func getQRCode(path string) (*gozxing.Result, error) {
	f, err := os.Open(path)
	if err != nil {
		return &gozxing.Result{}, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return &gozxing.Result{}, err
	}

	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return &gozxing.Result{}, err
	}

	qrReader := qrcode.NewQRCodeReader()
	result, err := qrReader.Decode(bmp, nil)
	if err != nil {
		return &gozxing.Result{}, err
	}

	return result, nil

}
