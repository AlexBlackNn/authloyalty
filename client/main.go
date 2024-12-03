package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func ImageToBase64(path string) []byte {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return []byte("")
	}
	var base64Encoding string
	mimeType := http.DetectContentType(bytes)
	switch mimeType {
	case "image/jpeg":
		base64Encoding += "data:image/jpeg;base64|"
	case "image/png":
		base64Encoding += "data:image/png;base64|"
	}
	base64Encoding += base64.StdEncoding.EncodeToString(bytes)

	fmt.Printf("%q\n", base64Encoding)

	data, err := base64.StdEncoding.DecodeString(strings.Split(base64Encoding, "|")[1])
	if err != nil {
		log.Fatal("error:", err)
	}

	err = ioutil.WriteFile("/home/alex/Dev/GolandYandex/authloyaltyminio/client/images1.jpeg", data, 0644)
	if err != nil {
		panic(err)
	}

	return data
}

func main() {
	print(ImageToBase64("/home/alex/Dev/GolandYandex/authloyaltyminio/client/images.jpeg"))
}
