package main

import (
	"log"
)

const (
	DrawTypeATP = "atp"
	DrawTypeWTA = "wta"
)

type HealthCheckRecorder interface {
	record(drawType, drawUrl, errMsg, token string)
}

type RealHealthCheckRecorder struct{}

func (r *RealHealthCheckRecorder) record(drawType, drawUrl, errMsg, token string) {
	addHealthCheck(drawType, drawUrl, errMsg, token)
}

func performHealthCheck(scraper Scraper, recorder HealthCheckRecorder, atpDraw, wtaDraw *DrawRecord, token string) {
	// Test ATP
	if atpDraw != nil {
		_, _, err := scrapeATP(scraper, *atpDraw)
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
			log.Printf("ATP health check failed: %v", err)
		} else {
			log.Println("ATP health check passed")
		}

		recorder.record(DrawTypeATP, atpDraw.Url, errMsg, token)
	} else {
		log.Println("No ATP draw found for health check")
	}

	// Test WTA
	if wtaDraw != nil {
		_, _, err := scrapeWTA(scraper, *wtaDraw)
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
			log.Printf("WTA health check failed: %v", err)
		} else {
			log.Println("WTA health check passed")
		}
		recorder.record(DrawTypeWTA, wtaDraw.Url, errMsg, token)
	} else {
		log.Println("No WTA draw found for health check")
	}
}
