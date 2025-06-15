import express from 'express'

const app = express()
const port = 8000

const images = []

app.use(express.json())

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

app.post('/api/message', async (req, res) => {
    const msg = req.body
    const result = await processMessage(msg)
    res.status(200).json({
        data: result
    })
})

app.use(express.static('client'))

app.listen(port, () => console.log(`Listening at http://localhost:${port}`))
