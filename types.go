package main

// Script types

type Slot struct {
	ID       string
	DrawID   string
	Round    int
	Position int
	Name     string
	Seed     string
	Sets     SetSlice
}

type Set struct {
	ID         string
	DrawSlotID string
	Number     int
	Games      int
	Tiebreak   int
}

type SlotSlice []Slot

func (ss *SlotSlice) add(s Slot) {
	*ss = append(*ss, s)
}

type SetSlice []Set

func (ss *SetSlice) add(s Set) {
	*ss = append(*ss, s)
}

// Pocketbase API types

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
	Size             int    `json:"size"`
}

type DrawRes struct {
	Page       int          `json:"page"`
	PerPage    int          `json:"perPage"`
	TotalItems int          `json:"totalItems"`
	TotalPages int          `json:"totalPages"`
	Items      []DrawRecord `json:"items"`
}

type SlotRecord struct {
	ID           string `json:"id"`
	DrawID       string `json:"draw_id"`
	Round        int    `json:"round"`
	Position     int    `json:"position"`
	Name         string `json:"name"`
	Seed         string `json:"seed"`
	Set1ID       string `json:"set1_id"`
	Set1Games    *int   `json:"set1_games"`
	Set1Tiebreak *int   `json:"set1_tiebreak"`
	Set2ID       string `json:"set2_id"`
	Set2Games    *int   `json:"set2_games"`
	Set2Tiebreak *int   `json:"set2_tiebreak"`
	Set3ID       string `json:"set3_id"`
	Set3Games    *int   `json:"set3_games"`
	Set3Tiebreak *int   `json:"set3_tiebreak"`
	Set4ID       string `json:"set4_id"`
	Set4Games    *int   `json:"set4_games"`
	Set4Tiebreak *int   `json:"set4_tiebreak"`
	Set5ID       string `json:"set5_id"`
	Set5Games    *int   `json:"set5_games"`
	Set5Tiebreak *int   `json:"set5_tiebreak"`
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

type CreateUpdateSetReq struct {
	DrawSlotID string `json:"draw_slot_id"`
	Number     int    `json:"number"`
	Games      int    `json:"games"`
	Tiebreak   int    `json:"tiebreak"`
}
