package main

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strings"
)

type Response struct {
	BytesString string `json:"bytes"`
}


type Unpack struct {
	Int             int32   `json:"int"`
	Uint            uint32  `json:"uint"`
	Short           int16   `json:"short"`
	Float           float32 `json:"float"`
	Double          float64 `json:"double"`
	BigEndianDouble float64 `json:"big_endian_double"`
}


func main() {
	res, err := http.Get("https://hackattic.com/challenges/help_me_unpack/problem?access_token=8f3a9cb1bc4572d2")
	if err != nil {
		fmt.Printf("error initiating GET request: %v", err)
		return
	}
	defer res.Body.Close()

	var b Response
	err = json.NewDecoder(res.Body).Decode(&b)
	if err != nil {
		fmt.Printf("error decoding json: %v", err)
		return
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(b.BytesString)
	if err != nil {
		log.Fatalf("decode error: %v", err)
		return
	}

	if len(decodedBytes) != 32 {
		fmt.Print("incorrect length of bytes")
		return
	}

	upack := Unpack{
		Int:             int32(binary.LittleEndian.Uint32(decodedBytes[:4])),
		Uint:            binary.LittleEndian.Uint32(decodedBytes[4:8]),
		Short:           int16(binary.LittleEndian.Uint16(decodedBytes[8:10])),
		Float:           math.Float32frombits(binary.LittleEndian.Uint32(decodedBytes[12:16])),
		Double:          math.Float64frombits(binary.LittleEndian.Uint64(decodedBytes[16:24])),
		BigEndianDouble: math.Float64frombits(binary.BigEndian.Uint64(decodedBytes[24:32])),
	}

	data, err := json.Marshal(upack)
	if err != nil {
		fmt.Printf("error marshalling json: %v", err)
		return
	}

	req, err := http.NewRequest("POST", "https://hackattic.com/challenges/help_me_unpack/solve?access_token=8f3a9cb1bc4572d2", strings.NewReader(string(data)))
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
