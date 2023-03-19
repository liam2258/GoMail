package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListenHandler(t *testing.T) {
	// Create a mock HTTP request.
	req, err := http.NewRequest("GET", "/listen", nil)
	if err != nil {
		t.Fatal(err)
	}

	//Create a ResponseRecorder to receive response back from http request
	res := httptest.NewRecorder()

	//Call function for test
	ListenHandler(res, req)

	//Check if the response status code is correct
	if res.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, res.Code)
	}

	//Check if the response body is correct
	expectedBody := "I'm listening!"
	if res.Body.String() != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, res.Body.String())
	}
}

func TestSendHandler(t *testing.T) {
	//Create raw email structure
	rawEmail := map[string]interface{}{
		"Name":    "Test",
		"Subject": "Test",
		"Contact": "1-800-TEST",
		"Message": "Hello this is a test!",
	}

	//Transforms raw email into JSON format
	email, err := json.Marshal(rawEmail)
	if err != nil {
		t.Fatal(err)
	}

	//Create http request
	req, err := http.NewRequest("POST", "/send", bytes.NewBuffer(email))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	// reate a ResponseRecorder to receive response back from http request
	res := httptest.NewRecorder()

	//Initialize client
	client := VerifyEmailClient()

	//Call function for test
	SendHandler(res, req, client)

	if res.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, res.Code)
	}

	// Check the response body is what we expect
	expected := "Email successfully sent!"
	if res.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			res.Body.String(), expected)
	}

	//Second SendHanler test, checks the response when sending a non-accepted method, in this case GET
	req, err = http.NewRequest("GET", "/send", bytes.NewBuffer(email))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")
	res = httptest.NewRecorder()

	//Call function for test2
	SendHandler(res, req, client)

	// Check if the response status code is correct
	if res.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status code %d, got %d", http.StatusMethodNotAllowed, res.Code)
	}
}
