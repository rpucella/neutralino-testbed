package main

import (
	"encoding/json"
	"encoding/base64"
	"github.com/google/uuid"
	"os"
	"bufio"
	"io"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	// "net/url"
	"os/signal"
)

// Mostly adapted from https://github.com/gorilla/websocket/blob/main/examples/echo/client.go

func main() {
	// Read connection information from Neutralino.
	reader := bufio.NewReader(os.Stdin)
	connInfoStr, err := reader.ReadString('\n')
	if err != nil && err != io.EOF{
		log.Fatal(fmt.Errorf("cannot read connection information: %w", err))
	}
	log.Println(connInfoStr)
	connInfo := make(map[string]string)
	if err := json.Unmarshal([]byte(connInfoStr), &connInfo); err != nil {
		log.Fatal(fmt.Errorf("cannot unmarshal json info: %w", err))
	}
	log.Println(connInfo)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	/*
        const NL_PORT = processInput.nlPort
        const NL_TOKEN = processInput.nlToken
        const NL_CTOKEN = processInput.nlConnectToken
        const NL_EXTID = processInput.nlExtensionId
        const NL_URL =  `ws://localhost:${NL_PORT}?extensionId=${NL_EXTID}&connectToken=${NL_CTOKEN}`
    */

	urlString := fmt.Sprintf("ws://localhost:%s?extensionId=%s&connectToken=%s",
		connInfo["nlPort"],
		connInfo["nlExtensionId"],
		connInfo["nlConnectToken"])

	//u := url.URL{Scheme: "ws", Host: *addr, Path: "/echo"}
	//log.Printf("connecting to %s", u.String())
	log.Printf("connecting to %s", urlString)

	c, _, err := websocket.DefaultDialer.Dial(urlString, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
			messageObj := make(map[string]interface{})
			if err := json.Unmarshal(message, &messageObj); err != nil {
				log.Println("cannot parse message:", err)
				continue
			}
			eventIfc, ok := messageObj["event"]
			if !ok {
				continue
			}
			event := eventIfc.(string)
			if event == "eventToExtension" {
				data := messageObj["data"].(map[string]interface{})
				callId := data["callId"].(float64)
				msgResult, err := processMessage(data)
				if err != nil {
					log.Println("cannot process message:", err)
					continue
				}
				result := make(map[string]interface{})
				result["id"] = uuid.NewString()
				result["method"] = "app.broadcast"
				result["accessToken"] = connInfo["nlToken"]
				dataResult := make(map[string]interface{})
				data2Result := make(map[string]interface{})
				dataResult["event"] = "eventFromExtension"
				data2Result["content"] = msgResult
				data2Result["callId"] = callId
				dataResult["data"] = data2Result
				result["data"] = dataResult
				obj, err := json.Marshal(result)
				if err != nil {
					log.Println("cannot marshal result:", err)
				}
				c.WriteMessage(websocket.BinaryMessage, obj)
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			}
			return
		}
	}
}

type image struct {
	name string
	content string // Base64 encoding
	mime string
}

var images []image = make([]image, 0)

func getImageDetails(key int) string {
	img := images[key]
	return fmt.Sprintf("data:%s;base64,%s", img.mime, img.content)
}

func getImages() []string {
	names := make([]string, 0)
	for _, img := range images {
		names = append(names, img.name)
	}
	return names
}

func addImage(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("cannot fetch image: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request status: %s", resp.StatusCode)
	}
	// get content-type
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("cannot read image: %w", err)
	}
	base64Str := base64.StdEncoding.EncodeToString(bodyBytes)
	contentType := resp.Header.Get("content-type")
	images = append(images, image{url, base64Str, contentType})
	return nil
}

func processMessage(message map[string]interface{}) (interface{}, error) {
	log.Println("processing message: ", message)
	mode := message["mode"].(string)
	switch(mode) {
	case "get-images":
		return getImages(), nil

	case "get-image":
		index := message["index"].(float64)
		return getImageDetails(int(index)), nil

	case "post-image":
		url := message["url"].(string)
		err := addImage(url)
		return "ok", err
	}
	return nil, fmt.Errorf("Unknown mode: %s", mode)
}
