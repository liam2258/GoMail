package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	sendinblue "github.com/sendinblue/APIv3-go-library/v2/lib"
)

// Const for setting Cors settings
const (
	options           string = "OPTIONS"
	allow_origin      string = "Access-Control-Allow-Origin"
	allow_methods     string = "Access-Control-Allow-Methods"
	allow_headers     string = "Access-Control-Allow-Headers"
	allow_credentials string = "Access-Control-Allow-Credentials"
	expose_headers    string = "Access-Control-Expose-Headers"
	credentials       string = "true"
	origin            string = "Origin"
	methods           string = "POST, GET, OPTIONS, PUT, DELETE, HEAD, PATCH"

	headers string = "Access-Control-Allow-Origin, Accept, Accept-Encoding, Authorization, Content-Length, Content-Type, X-CSRF-Token"
)

// Adds Cors to HTTP handler
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set allow origin to match origin of our request or fall back to *
		if o := r.Header.Get(origin); o != "" {
			w.Header().Set(allow_origin, o)
		} else {
			w.Header().Set(allow_origin, "*")
		}

		// Set other headers
		w.Header().Set(allow_headers, headers)
		w.Header().Set(allow_methods, methods)
		w.Header().Set(allow_credentials, credentials)
		w.Header().Set(expose_headers, headers)

		// If this was preflight options request, write empty ok response and return
		if r.Method == options {
			w.WriteHeader(http.StatusOK)
			w.Write(nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Establish connection to email service sendinblue
func VerifyEmailClient() *sendinblue.APIClient {
	var ctx context.Context
	cfg := sendinblue.NewConfiguration()
	//Configure API key authorization: api-key
	cfg.AddDefaultHeader("api-key", os.Getenv("API_KEY"))
	//Configure API key authorization: partner-key
	cfg.AddDefaultHeader("partner-key", os.Getenv("API_KEY"))

	client := sendinblue.NewAPIClient(cfg)
	_, _, err := client.AccountApi.GetAccount(ctx)
	if err != nil {
		fmt.Println("Error when calling AccountApi->get_account: ", err.Error())
		panic(err)
	}
	fmt.Println("Successfully connected to client!")
	return client
}

// Creates the email using the given data
func CreateEmail(name, subject, contact, message string) sendinblue.SendSmtpEmail {
	sender := sendinblue.SendSmtpEmailSender{
		Name:  name,
		Email: os.Getenv("SEND_EMAIL"),
	}
	recipient := []sendinblue.SendSmtpEmailTo{{
		Email: os.Getenv("RECEIVE_EMAIL"),
		Name:  "Liam",
	}}

	content := "<html><body><p>" + message + "</p></body></html>"

	email := sendinblue.SendSmtpEmail{Sender: &sender, To: recipient, Subject: subject, HtmlContent: content}

	return email
}

// Sends out the email
func SendEmail(client *sendinblue.APIClient, email sendinblue.SendSmtpEmail) error {
	var ctx context.Context

	_, _, err := client.TransactionalEmailsApi.SendTransacEmail(ctx, email)

	return err
}

// Handles the listen request
func ListenHandler(write http.ResponseWriter, read *http.Request) {
	write.WriteHeader(http.StatusOK)
	fmt.Fprint(write, "I'm listening!")
}

// Handles the send request
func SendHandler(write http.ResponseWriter, read *http.Request, client *sendinblue.APIClient) {

	//Prevents unsupported methods
	if read.Method != "POST" {
		http.Error(write, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	//Transforms the request body into a byte slice and returns if there is an error during this process
	body, err := io.ReadAll(read.Body)
	if err != nil {
		http.Error(write, err.Error(), http.StatusBadRequest)
		return
	}

	//Turns the body into a map structure
	var data map[string]interface{}

	//Turns the JSON map structure into a normal string map structure
	if err := json.Unmarshal(body, &data); err != nil {
		http.Error(write, err.Error(), http.StatusBadRequest)
		return
	}

	//Checks the format of the sent JSON object
	if len(data) != 4 {
		http.Error(write, "Invalid format", http.StatusBadRequest)
	}

	//Creates the email using the recieved data
	email := CreateEmail(data["Name"].(string), data["Subject"].(string), data["Contact"].(string), data["Message"].(string))

	//Sends the email using the smtp client
	emailErr := SendEmail(client, email)

	// Send a response back to the requester
	if emailErr != nil {
		write.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(write, "Email not sent, service error")
	} else {
		write.WriteHeader(http.StatusOK)
		fmt.Fprint(write, "Email successfully sent!")
	}
}

func main() {
	//Set port to 8080 if environment variable isn't set
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	//Initializes client for sendinblue smtp service
	client := VerifyEmailClient()

	//Handles listen requests to make sure the server is up
	mux.HandleFunc("/listen", ListenHandler)

	//Handles send request to send out email after recieving contents in JSON format
	mux.HandleFunc("/send", func(write http.ResponseWriter, read *http.Request) {
		SendHandler(write, read, client)
	})

	//Sets server to begin listening for requests
	fmt.Println("Listening on port ", port)
	http.ListenAndServe(":"+port, CORS(mux))
}
