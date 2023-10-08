package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type UserRecord struct {
	ID              string `json:"id"`
	Created         string `json:"created"`
	Updated         string `json:"updated"`
	Username        string `json:"username"`
	Email           string `json:"email"`
	Verified        bool   `json:"verified"`
	EmailVisibility bool   `json:"emailVisibility"`
}

type UserAuthRes struct {
	Token  string     `json:"token"`
	Record UserRecord `json:"record"`
}

type DrawRecord struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Event            string `json:"event"`
	Year             int    `json:"year"`
	Url              string `json:"url"`
	Start_Date       string `json:"start_date"`
	End_Date         string `json:"end_date"`
	Prediction_Close string `json:"prediction_close"`
}

type DrawRes struct {
	Page       int          `json:"page"`
	PerPage    int          `json:"perPage"`
	TotalItems int          `json:"totalItems"`
	TotalPages int          `json:"totalPages"`
	Items      []DrawRecord `json:"items"`
}

type SlotRecord struct {
	ID       string `json:"id"`
	DrawID   string `json:"draw_id"`
	Round    int    `json:"round"`
	Position int    `json:"position"`
	Name     string `json:"name"`
	Seed     string `json:"seed"`
}

type SlotRes struct {
	Page       int          `json:"page"`
	PerPage    int          `json:"perPage"`
	TotalPages int          `json:"totalPages"`
	TotalItems int          `json:"totalItems"`
	Items      []SlotRecord `json:"items"`
}

type CreateUpdateSlotReq struct {
	DrawID   string `json:"draw_id"`
	Round    int    `json:"round"`
	Position int    `json:"position"`
	Name     string `json:"name"`
	Seed     string `json:"seed"`
}

type CreateSlotRes struct {
	ID       string `json:"id"`
	DrawID   string `json:"draw_id"`
	Name     string `json:"name"`
	Seed     int    `json:"seed"`
	Round    int    `json:"round"`
	Position int    `json:"position"`
}

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
		fmt.Println(err)
		return ""
	}
	defer res.Body.Close()

	userAuthRes := &UserAuthRes{}
	fmt.Println("Auth request status:", res.Status)
	derr := json.NewDecoder(res.Body).Decode(userAuthRes)
	if derr != nil {
		fmt.Println(err)
		return ""
	}

	return userAuthRes.Token
}

func getDraws(token string) []DrawRecord {
	today := time.Now().UTC().Format("2006-01-02")
	url := fmt.Sprintf(`%s/api/collections/draw/records?filter=(end_date>="%s")&fields=id,name,event,year,url,start_date,end_date,prediction_close`, os.Getenv("BASE_URL"), today)

	res, err := makeHTTPRequest("GET", url, token, nil)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer res.Body.Close()

	drawRes := &DrawRes{}
	derr := json.NewDecoder(res.Body).Decode(drawRes)
	if derr != nil {
		fmt.Println(derr)
		return nil
	}

	return drawRes.Items
}

func getSlots(drawId string, token string) []SlotRecord {
	url := fmt.Sprintf(`%s/api/collections/draw_slot/records?perPage=255&filter=(draw_id="%s")&skipTotal=true`, os.Getenv("BASE_URL"), drawId)

	res, err := makeHTTPRequest("GET", url, token, nil)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()

	slotRes := &SlotRes{}
	derr := json.NewDecoder(res.Body).Decode(slotRes)
	if derr != nil {
		fmt.Println(derr)
		return nil
	}

	return slotRes.Items
}

func postSlots(slots slotSlice, token string) {
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
			fmt.Println(err)
		}
		defer res.Body.Close()

		fmt.Println(res.Status, "added", slot)
	}
}

func updateSlots(slots slotSlice, token string) {
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
			fmt.Println(err)
		}
		defer res.Body.Close()

		fmt.Println(res.Status, "updated", slot)
	}
}
