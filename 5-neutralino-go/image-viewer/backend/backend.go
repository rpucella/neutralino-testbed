package main

import (
	"encoding/json"
	"os"
	"bufio"
	"io"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
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
			result, err := processMessage(string(message))
			log.Println(result)
			if err != nil {
				log.Printf("cannot process message:", err)
				return
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


func processMessage(message string) (string, error) {
	log.Println(message)
	return "", nil
}


/*
type connInfo struct {
	Port string "json:nlPort"
	Token string "json:nlToken"
	CToken string "json:nlConnectToken"
	ExtId string "json:nlExtensionId"
}
*/

/*
const processInput = JSON.parse(fs.readFileSync(process.stdin.fd, 'utf-8'))
const NL_PORT = processInput.nlPort
const NL_TOKEN = processInput.nlToken
const NL_CTOKEN = processInput.nlConnectToken
const NL_EXTID = processInput.nlExtensionId
const NL_URL =  `ws://localhost:${NL_PORT}?extensionId=${NL_EXTID}&connectToken=${NL_CTOKEN}`
*/



/*
const images = []

class Storage {
    constructor() {
        this.images = []
    }

    readImage(index) {
        return this.images[index]
    }

    readImageNames() {
        return this.images.map(img => img.name)
    }

    createImage(name, content, contentType) {
        this.images.push({
            name: name,
            content: content,
            mime: contentType
        })
    }
}


class Controller {
    constructor() {
        this.store = new Storage()
    }

    getImageDetails(key) {
        const img = this.store.readImage(key)
        return `data:${img.mime};base64,${img.content}`
    }

    getImages() {
        return this.store.readImageNames()
    }

    async addImage(url) {
        const response = await fetch(url)
        const contentType = response.headers.get('content-type')
        const abuffer = await response.arrayBuffer()
        const base64Image = Buffer.from(abuffer).toString('base64')
        this.store.createImage(url, base64Image, contentType)
    }
}

const controller = new Controller()

async function processMessage(msg) {
    switch(msg.mode) {
    case "get-image":
        return controller.getImageDetails(msg.index)
        break

    case "get-images":
        return controller.getImages()
        break

    case "post-image":
        await controller.addImage(msg.url)
        return "ok"
    }
    console.log(`Error - unknown message type ${msg.mode}`)
    return "unknown message type"
}

*/

/*

import WebSocket from 'ws'

const client = new WebSocket(NL_URL)

client.on('error', (error) => {
    console.log(`Connection error!`)
    console.dir(error, {depth:null})
})
client.on('open', () => console.log("Connected"))
client.on('close', (code, reason) => {
  console.log(`WebSocket closed: ${code} - ${reason}`);
  process.exit()
})
client.on('message', async (evt) => {
  console.log("Event = ", evt)
  const evtData = evt.toString('utf-8')
  const { event, data } = JSON.parse(evtData)

  if (event === "eventToExtension") {
    const callId = data.callId
    const result = await processMessage(data)
    client.send(JSON.stringify({
      method: "app.broadcast",
      accessToken: NL_TOKEN,
      data: {
        event: "eventFromExtension",
        data: {content: result, callId}
      }
    }))
 }
})

*/
