package main

import (
	"log"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

type MockHealthCheckRecorder struct {
	Calls []HealthCheckCall
}

type HealthCheckCall struct {
	DrawType string
	DrawUrl  string
	ErrMsg   string
	Token    string
}

func (m *MockHealthCheckRecorder) Record(drawType, drawUrl, errMsg, token string) {
	m.Calls = append(m.Calls, HealthCheckCall{
		DrawType: drawType,
		DrawUrl:  drawUrl,
		ErrMsg:   errMsg,
		Token:    token,
	})
}

func TestPerformHealthCheck(t *testing.T) {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file,", err)
	}

	// // Set up mock environment
	// os.Setenv("BASE_URL", "http://localhost:8090")

	t.Run("Health check with mock scraper and recorder", func(t *testing.T) {
		scraper := &MockScraper{}
		recorder := &MockHealthCheckRecorder{}
		token := "test_token"

		// Create mock draws
		atpDraw := &DrawRecord{
			ID:               "test_atp_id",
			Name:             "Australian Open",
			Event:            "Men's Singles",
			Year:             2025,
			Url:              "https://www.atptour.com/en/scores/current/australian-open/580/draws",
			Start_Date:       "2025-01-12 12:00:00.000",
			End_Date:         "2025-01-26 12:00:00.000",
			Prediction_Close: "2025-01-19 12:00:00.000",
			Size:             128,
		}

		wtaDraw := &DrawRecord{
			ID:               "test_wta_id",
			Name:             "Australian Open",
			Event:            "Women's Singles",
			Year:             2025,
			Url:              "https://www.live-tennis.eu/en/wta-singles-draws",
			Start_Date:       "2025-01-12 12:00:00.000",
			End_Date:         "2025-01-26 12:00:00.000",
			Prediction_Close: "2025-01-19 12:00:00.000",
			Size:             128,
		}

		assert := assert.New(t)

		// Call performHealthCheck with mock draws
		performHealthCheck(scraper, recorder, atpDraw, wtaDraw, token)

		// Verify recorder was called twice
		assert.Equal(2, len(recorder.Calls))

		// Verify ATP call
		assert.Equal(DrawTypeATP, recorder.Calls[0].DrawType)
		assert.Equal("https://www.atptour.com/en/scores/current/australian-open/580/draws", recorder.Calls[0].DrawUrl)
		assert.Equal("", recorder.Calls[0].ErrMsg)
		assert.Equal(token, recorder.Calls[0].Token)

		// Verify WTA call
		assert.Equal(DrawTypeWTA, recorder.Calls[1].DrawType)
		assert.Equal("https://www.live-tennis.eu/en/wta-singles-draws", recorder.Calls[1].DrawUrl)
		assert.Equal("", recorder.Calls[1].ErrMsg)
		assert.Equal(token, recorder.Calls[1].Token)
	})

	t.Run("Health check with nil draws", func(t *testing.T) {
		scraper := &MockScraper{}
		recorder := &MockHealthCheckRecorder{}
		token := "test_token"

		assert := assert.New(t)

		// Call with nil draws
		performHealthCheck(scraper, recorder, nil, nil, token)

		// Verify recorder was not called
		assert.Equal(0, len(recorder.Calls))
	})

	t.Run("Health check records errors", func(t *testing.T) {
		scraper := &MockScraper{}
		recorder := &MockHealthCheckRecorder{}
		token := "test_token"

		// Test with invalid URL that MockScraper doesn't recognize
		invalidDraw := &DrawRecord{
			ID:               "test_invalid_id",
			Name:             "Invalid Tournament",
			Event:            "Men's Singles",
			Year:             2025,
			Url:              "https://invalid.example.com/draws",
			Start_Date:       "2025-01-12 12:00:00.000",
			End_Date:         "2025-01-26 12:00:00.000",
			Prediction_Close: "2025-01-19 12:00:00.000",
			Size:             128,
		}

		assert := assert.New(t)

		// Call performHealthCheck with invalid draw
		performHealthCheck(scraper, recorder, invalidDraw, nil, token)

		// Verify error was recorded
		assert.Equal(1, len(recorder.Calls))
		assert.Equal(DrawTypeATP, recorder.Calls[0].DrawType)
		assert.Contains(recorder.Calls[0].ErrMsg, "unknown URL")
	})
}
