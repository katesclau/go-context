package main

import (
  "context"
  "fmt"
  "net/http"
  "encoding/json"
  "io/ioutil"
  "bytes"
  "time"
  "os"

  "github.com/joho/godotenv"
)

// Define the Go structs
type VerificationMethod struct {
  ID                 string `json:"id"`
  Type               string `json:"type"`
  Controller         string `json:"controller"`
  PublicKeyMultibase string `json:"publicKeyMultibase"`
}

type Service struct {
  ID              string `json:"id"`
  Type            string `json:"type"`
  ServiceEndpoint string `json:"serviceEndpoint"`
}

type DidDoc struct {
  Context            []string             `json:"@context"`
  ID                 string               `json:"id"`
  AlsoKnownAs        []string             `json:"alsoKnownAs"`
  VerificationMethod []VerificationMethod `json:"verificationMethod"`
  Service            []Service            `json:"service"`
}

type Response struct {
  Did             string `json:"did"`
  DidDoc          DidDoc `json:"didDoc"`
  Handle          string `json:"handle"`
  Email           string `json:"email"`
  EmailConfirmed  bool   `json:"emailConfirmed"`
  EmailAuthFactor bool   `json:"emailAuthFactor"`
  AccessJwt       string `json:"accessJwt"`
  RefreshJwt      string `json:"refreshJwt"`
  Active          bool   `json:"active"`
}

type bytesOrError struct {
  data []byte
  err  error
}

/** fetchAPI
  * 1. Create a new request with the URL
  * 2. Set the headers from the context
  * 3. Make the request
  * 4. Read the response body
  * 5. Send the response body to the results channel
  */
func fetchAPI(ctx context.Context, url string, results chan<- bytesOrError) {
  req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
  if err != nil {
    results <- bytesOrError{[]byte{}, err}
    return
  }

  if headers, ok := ctx.Value("headers").(map[string]string); ok {
    for key, value := range headers {
      req.Header.Set(key, value)
    }
  }

  client := http.DefaultClient
  resp, err := client.Do(req)
  if err != nil {
    results <- bytesOrError{[]byte{}, err}
    return
  }
  defer resp.Body.Close()

  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    results <- bytesOrError{[]byte{}, err}
    return 
  }

  results <- bytesOrError{body, nil}
}

/** fetchSession
  * 1. Create a map of the handle and password
  * 2. Marshal the data into JSON
  * 3. Create a new request with the URL and data
  * 4. Set the content type to application/json
  * 5. Make the request
  * 6. Read the response body
  * 7. Unmarshal the response body into a Response struct
  */
func fetchSession(ctx context.Context, url string, handle string, password string) (Response, error) {
  data := map[string]string{"identifier": handle, "password": password}

  jsonData, err := json.Marshal(data)
  if err != nil {
    return Response{}, fmt.Errorf("Error marshalling data for %s: %s", url, err.Error())
  }

  req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
  if err != nil {
    return Response{}, fmt.Errorf("Error creating request for %s: %s", url, err.Error())
  }

  req.Header.Set("Content-Type", "application/json")

  client := http.DefaultClient
  resp, err := client.Do(req)
  if err != nil {
    return Response{}, fmt.Errorf("Error making request to %s: %s", url, err.Error())
  }
  defer resp.Body.Close()

  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return Response{}, fmt.Errorf("Error reading response body from %s: %s", url, err.Error())
  }

  var response Response
  err = json.Unmarshal([]byte(body), &response)
  if err != nil {
    return Response{}, fmt.Errorf("Error unmarshalling response from %s: %s", url, err.Error())
  }

  return response, nil
}

/** MAIN 
  * 1. Create a context with a timeout of 5 seconds
  * 2. Define the URLs to fetch
  * 3. Create a channel to store the results
  * 4. Fetch the session
  * 5. Create a map of headers to pass to the context
  * 6. Loop through the URLs and fetch the data
  * 7. Print the results
  */
func main() {
  err := godotenv.Load()
  if err != nil {
    fmt.Println("Error loading .env file")
    return
  }

  handle := os.Getenv("HANDLE")
  password := os.Getenv("PASSWORD")

  ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
  defer cancel()

  urls := []string{
    fmt.Sprintf("https://bsky.social/xrpc/app.bsky.actor.getProfile?actor=%s", handle),
    fmt.Sprintf("https://bsky.social/xrpc/app.bsky.feed.getActorFeeds?actor=%s", handle),
  }

  results := make(chan bytesOrError, len(urls))

  res, err := fetchSession(ctx, "https://bsky.social/xrpc/com.atproto.server.createSession", handle, password)
  if err != nil {
    fmt.Println(err)
    return
  }
  fmt.Printf("Session: %v\n", res)
  headers := map[string]string{"Authorization": fmt.Sprintf("Bearer %s", res.AccessJwt)}
  ctx = context.WithValue(ctx, "headers", headers)

  for _, url := range urls {
    go fetchAPI(ctx, url, results)
  }

  for range urls {
    res := <-results
    if res.err != nil {
      fmt.Println(res.err)
      return
    }

    // prettyJSON, err := json.MarshalIndent(res.data, "", "  ")
    // if err != nil {
    //   fmt.Println(err)
    //   return
    // }

    var prettyJSON bytes.Buffer
    err := json.Indent(&prettyJSON, res.data, "", "  ")
    if err != nil {
        fmt.Println("Error formatting JSON:", err)
        return
    }

    fmt.Println("VVVVVVVVVVVVVVVV")
    fmt.Println("",prettyJSON.String())
    fmt.Println("----------------")
  }
}

