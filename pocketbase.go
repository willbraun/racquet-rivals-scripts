package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func makeHTTPRequest(method, url, token string, requestData interface{}) (*http.Response, error) {
	body, err := json.Marshal(requestData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	if token != "" {
		req.Header.Add("Authorization", token)
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	return client.Do(req)
}

func login() string {
	url := fmt.Sprintf(`%s/api/collections/user/auth-with-password`, os.Getenv("BASE_URL"))

	requestData := struct {
		Identity string `json:"identity"`
		Password string `json:"password"`
	}{
		Identity: os.Getenv("SCRIPT_USER_USERNAME"),
		Password: os.Getenv("SCRIPT_USER_PASSWORD"),
	}

	res, err := makeHTTPRequest("POST", url, "", requestData)
	if err != nil {
		log.Println(err)
		return ""
	}
	defer res.Body.Close()

	userAuthRes := &UserAuthRes{}
	derr := json.NewDecoder(res.Body).Decode(userAuthRes)
	if derr != nil {
		log.Println(derr)
		return ""
	}

	return userAuthRes.Token
}

func getDraws(token string) []DrawRecord {
	today := time.Now().UTC().Format("2006-01-02")
	url := fmt.Sprintf(`%s/api/collections/draw/records?filter=(end_date>="%s")&fields=id,name,event,year,url,start_date,end_date,prediction_close,size`, os.Getenv("BASE_URL"), today)

	res, err := makeHTTPRequest("GET", url, token, nil)
	if err != nil {
		log.Println(err)
		return nil
	}
	defer res.Body.Close()

	drawRes := &DrawRes{}
	derr := json.NewDecoder(res.Body).Decode(drawRes)
	if derr != nil {
		log.Println(derr)
		return nil
	}

	return drawRes.Items
}

func getSlots(drawId string, token string) SlotSlice {
	url := fmt.Sprintf(`%s/api/collections/slots_with_scores/records?perPage=255&filter=(draw_id="%s")&skipTotal=true`, os.Getenv("BASE_URL"), drawId)

	res, err := makeHTTPRequest("GET", url, token, nil)
	if err != nil {
		log.Println(err)
	}
	defer res.Body.Close()

	slotRes := &SlotRes{}
	derr := json.NewDecoder(res.Body).Decode(slotRes)
	if derr != nil {
		log.Println(derr)
		return nil
	}

	return toSlotSlice(slotRes.Items)
}

func postSlots(slots SlotSlice, token string) {
	if len(slots) == 0 {
		return
	}

	url := fmt.Sprintf(`%s/api/collections/draw_slot/records`, os.Getenv("BASE_URL"))

	for _, slot := range slots {
		requestData := CreateUpdateSlotReq{
			DrawID:   slot.DrawID,
			Round:    slot.Round,
			Position: slot.Position,
			Name:     slot.Name,
			Seed:     slot.Seed,
		}
		res, err := makeHTTPRequest("POST", url, token, requestData)
		if err != nil {
			log.Println(err)
		}
		defer res.Body.Close()

		printWithTimestamp(res.Status, "added slot", slot)
	}
}

func updateSlots(slots SlotSlice, token string) {
	if len(slots) == 0 {
		return
	}

	for _, slot := range slots {
		url := fmt.Sprintf(`%s/api/collections/draw_slot/records/%s`, os.Getenv("BASE_URL"), slot.ID)
		requestData := CreateUpdateSlotReq{
			DrawID:   slot.DrawID,
			Round:    slot.Round,
			Position: slot.Position,
			Name:     slot.Name,
			Seed:     slot.Seed,
		}
		res, err := makeHTTPRequest("PATCH", url, token, requestData)
		if err != nil {
			log.Println(err)
		}
		defer res.Body.Close()

		printWithTimestamp(res.Status, "updated slot", slot)
	}
}

func postSets(setScores SetSlice, token string) {
	if len(setScores) == 0 {
		return
	}

	url := fmt.Sprintf(`%s/api/collections/set_score/records`, os.Getenv("BASE_URL"))

	for _, setScore := range setScores {
		requestData := CreateUpdateSetReq{
			DrawSlotID: setScore.DrawSlotID,
			Number:     setScore.Number,
			Games:      setScore.Games,
			Tiebreak:   setScore.Tiebreak,
		}
		res, err := makeHTTPRequest("POST", url, token, requestData)
		if err != nil {
			log.Println(err)
		}
		defer res.Body.Close()

		printWithTimestamp(res.Status, "added set", setScore)
	}
}

func updateSets(setScores SetSlice, token string) {
	if len(setScores) == 0 {
		return
	}

	for _, setScore := range setScores {
		url := fmt.Sprintf(`%s/api/collections/set_score/records/%s`, os.Getenv("BASE_URL"), setScore.ID)
		requestData := CreateUpdateSetReq{
			DrawSlotID: setScore.DrawSlotID,
			Number:     setScore.Number,
			Games:      setScore.Games,
			Tiebreak:   setScore.Tiebreak,
		}
		res, err := makeHTTPRequest("PATCH", url, token, requestData)
		if err != nil {
			log.Println(err)
		}
		defer res.Body.Close()

		printWithTimestamp(res.Status, "updated set", setScore)
	}
}
