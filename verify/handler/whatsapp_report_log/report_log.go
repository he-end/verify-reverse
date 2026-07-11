package report

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"log"
// 	"net/http"
// 	"sync"
// 	"wa/app/conf"
// 	"wa/app/handler/models"
// )

// type ModelData struct {
// 	LastIndex int    `json:"lastIndex"`
// 	Token     string `json:"token"`
// }

// func ReportLog(w http.ResponseWriter, r *http.Request) {
// 	if r.Method == "GET" {
// 		w.WriteHeader(http.StatusMethodNotAllowed)
// 		w.Write([]byte("method not allowed"))
// 	}
// 	// verify
// 	model := ModelData{}
// 	body_client := r.Body
// 	decode := json.NewDecoder(body_client)
// 	decode.DisallowUnknownFields()
// 	decodeErr := decode.Decode(&model)
// 	if decodeErr != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		w.Write([]byte(`{"status":"params invalid"}`))
// 		return
// 	}
// 	if !conf.Verify(model.Token) {
// 		w.WriteHeader(http.StatusUnauthorized)
// 		w.Write([]byte(`"status":"unauthorized"`))
// 		return
// 	}

// 	var wg sync.WaitGroup
// 	wg.Add(1)
// 	go func() {
// 		defer wg.Done()
// 		w.Write([]byte(`{"status":"200OK"}`))
// 	}()
// 	// report to 6285869144649
// 	urls := fmt.Sprintf("%v/%v/messages", utils.Utils.BaseURLGraphAPI, utils.Utils.PhoneNumberID)

// 	to := "6285869144649"
// 	token := utils.Utils.TokenWhatsApp
// 	typeText := models.TextBody{Body: fmt.Sprintf("%v", model.LastIndex)}
// 	messageModels := models.SendMessageModels{
// 		MessagingProduct: "whatsapp",
// 		To:               to,
// 		Type:             "text",
// 		Text:             &typeText,
// 	}
// 	body, _ := json.Marshal(messageModels)
// 	// log.Println(string(body))
// 	req, _ := http.NewRequest("POST", urls, bytes.NewBuffer(body))
// 	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
// 	req.Header.Set("Content-Type", "application/json")
// 	client := http.Client{}
// 	resp, _ := client.Do(req)
// 	readResp, _ := io.ReadAll(resp.Body)
// 	log.Println(string(readResp))
// 	wg.Wait()

// }
