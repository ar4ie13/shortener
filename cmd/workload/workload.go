// Workload is used to generate workload for service benchmarking and profiling
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	generateLength      = 10
	randGenerateSymbols = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

var endpoint = "http://localhost:8080"

type jsonRequest struct {
	URL string `json:"url"`
}

type jsonRequestBatch struct {
	UUID uuid.UUID `json:"correlation_id"`
	URL  string    `json:"original_url"`
}

func main() {

	for i := 0; i < 1000; i++ {
		testRoot()
	}
	for i := 0; i < 1000; i++ {

		testJSON()
	}
	for i := 0; i < 1000; i++ {
		testJSONBatch()
	}

	time.Sleep(10 * time.Second)
}

func generateURL(length int) string {
	generatedURL := make([]byte, length)
	for i := range generatedURL {
		generatedURL[i] = randGenerateSymbols[rand.Intn(len(randGenerateSymbols))]
	}
	finalURL := fmt.Sprintf("http://%s.com", string(generatedURL))
	return finalURL

}

func testJSONBatch() {
	jsonReqs := []jsonRequestBatch{}
	for i := 0; i < 1000; i++ {
		jsonReqs = append(jsonReqs, jsonRequestBatch{
			uuid.New(),
			generateURL(generateLength),
		})
	}

	req, err := json.Marshal(jsonReqs)
	if err != nil {
		log.Fatal(err)
	}
	reader := bytes.NewReader(req)
	// добавляем HTTP-клиент
	client := &http.Client{}
	// пишем запрос
	// запрос методом POST должен, помимо заголовков, содержать тело
	path := endpoint + "/api/shorten/batch"
	request, err := http.NewRequest(http.MethodPost, path, reader)
	if err != nil {
		panic(err)
	}
	// в заголовках запроса указываем кодировку
	request.Header.Add("Content-Type", "application/json")
	// отправляем запрос и получаем ответ
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	// выводим код ответа
	fmt.Printf("JSON BATCH: Статус-код %v\n", response.Status)
	defer response.Body.Close()
}

func testJSON() {
	jsonReq := jsonRequest{URL: generateURL(generateLength)}

	req, err := json.Marshal(jsonReq)
	if err != nil {
		log.Fatal(err)
	}
	reader := bytes.NewReader(req)
	// добавляем HTTP-клиент
	client := &http.Client{}
	// пишем запрос
	// запрос методом POST должен, помимо заголовков, содержать тело
	path := endpoint + "/api/shorten"
	request, err := http.NewRequest(http.MethodPost, path, reader)
	if err != nil {
		panic(err)
	}
	// в заголовках запроса указываем кодировку
	request.Header.Add("Content-Type", "application/json")
	// отправляем запрос и получаем ответ
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	// выводим код ответа
	fmt.Printf("JSON: Статус-код %v\n", response.Status)
	defer response.Body.Close()
}

func testRoot() {
	reader := strings.NewReader(generateURL(generateLength))
	// добавляем HTTP-клиент
	client := &http.Client{}
	// пишем запрос
	// запрос методом POST должен, помимо заголовков, содержать тело
	path := endpoint + "/"
	request, err := http.NewRequest(http.MethodPost, path, reader)
	if err != nil {
		panic(err)
	}
	// в заголовках запроса указываем кодировку
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	// отправляем запрос и получаем ответ
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	// выводим код ответа
	fmt.Printf("ROOT: Статус-код %v\n", response.Status)
	defer response.Body.Close()
}
