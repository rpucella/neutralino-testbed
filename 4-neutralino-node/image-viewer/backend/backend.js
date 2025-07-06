import { v4 as uuidV4 } from "uuid"

import process from "process"
import fs from "fs"
const processInput = JSON.parse(fs.readFileSync(process.stdin.fd, 'utf-8'))
const NL_PORT = processInput.nlPort
const NL_TOKEN = processInput.nlToken
const NL_CTOKEN = processInput.nlConnectToken
const NL_EXTID = processInput.nlExtensionId
const NL_URL =  `ws://localhost:${NL_PORT}?extensionId=${NL_EXTID}&connectToken=${NL_CTOKEN}`

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

// Routes.

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
  const evtData = evt.toString('utf-8')
  console.log("Event = ", evtData)
  const { event, data } = JSON.parse(evtData)
  if (event === "eventToExtension") {
    const callId = data.callId
    const result = await processMessage(data)
    client.send(JSON.stringify({
      id: uuidV4(),
      method: "app.broadcast",
      accessToken: NL_TOKEN,
      data: {
        event: "eventFromExtension",
        data: {content: result, callId}
      }
    }))
 }
})
