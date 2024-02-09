package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
)

// IncomingRequest represents the structure of incoming JSON requests
type IncomingRequest struct {
    Ev  string `json:"ev"`
    Et  string `json:"et"`
    ID  string `json:"id"`
    UID string `json:"uid"`
    Mid string `json:"mid"`
    T   string `json:"t"`
    P   string `json:"p"`
    L   string `json:"l"`
    SC  string `json:"sc"`
    Atr []struct {
        Key   string `json:"atrk"`
        Value string `json:"atrv"`
        Type  string `json:"atrt"`
    } `json:"-"`
    Uatr []struct {
        Key   string `json:"uatrk"`
        Value string `json:"uatrv"`
        Type  string `json:"uatrt"`
    } `json:"-"`
}

// TransformedRequest represents the structure of transformed JSON requests
type TransformedRequest struct {
    Event          string                 `json:"event"`
    EventType      string                 `json:"event_type"`
    AppID          string                 `json:"app_id"`
    UserID         string                 `json:"user_id"`
    MessageID      string                 `json:"message_id"`
    PageTitle      string                 `json:"page_title"`
    PageURL        string                 `json:"page_url"`
    BrowserLanguage string                 `json:"browser_language"`
    ScreenSize     string                 `json:"screen_size"`
    Attributes     map[string]interface{} `json:"attributes"`
    Traits         map[string]interface{} `json:"traits"`
}

func main() {
    // Create a channel to send incoming requests to the worker
    requests := make(chan IncomingRequest)

    // Start worker goroutine
    go worker(requests)

    // HTTP server to receive incoming requests
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        var incomingRequest IncomingRequest
        if err := json.NewDecoder(r.Body).Decode(&incomingRequest); err != nil {
            http.Error(w, "Invalid request body", http.StatusBadRequest)
            return
        }
        defer r.Body.Close()

        // Send incoming request to the worker via channel
        requests <- incomingRequest

        // Respond to the client
        w.WriteHeader(http.StatusOK)
        fmt.Fprintf(w, "Request received and sent to worker")
    })

    fmt.Println("Server listening on port 8080...")
    http.ListenAndServe(":8080", nil)
}

func worker(requests <-chan IncomingRequest) {
    for incomingRequest := range requests {
        go func(incoming IncomingRequest) {
            transformedRequest := TransformRequest(incoming)
            SendToWebhook(transformedRequest)
        }(incomingRequest)
    }
}

func TransformRequest(incoming IncomingRequest) TransformedRequest {
    transformed := TransformedRequest{
        Event:          incoming.Ev,
        EventType:      incoming.Et,
        AppID:          incoming.ID,
        UserID:         incoming.UID,
        MessageID:      incoming.Mid,
        PageTitle:      incoming.T,
        PageURL:        incoming.P,
        BrowserLanguage: incoming.L,
        ScreenSize:     incoming.SC,
        Attributes:     make(map[string]interface{}),
        Traits:         make(map[string]interface{}),
    }

    for _, attr := range incoming.Atr {
        transformed.Attributes[attr.Key] = map[string]interface{}{
            "value": attr.Value,
            "type":  attr.Type,
        }
    }

    for _, uatr := range incoming.Uatr {
        transformed.Traits[uatr.Key] = map[string]interface{}{
            "value": uatr.Value,
            "type":  uatr.Type,
        }
    }

    return transformed
}

func SendToWebhook(transformed TransformedRequest) {
    // Convert transformed to JSON
    jsonData, err := json.Marshal(transformed)
    if err != nil {
        fmt.Println("Error marshaling JSON:", err)
        return
    }

    // Send jsonData to webhook endpoint
    _, err = http.Post("https://webhook.site/", "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        fmt.Println("Error sending request to webhook:", err)
        return
    }
    fmt.Println("Request sent to webhook successfully")
}

