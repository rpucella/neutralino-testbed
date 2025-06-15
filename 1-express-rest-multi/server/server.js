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

// Routes.

app.post('/api/images', async (req, res) => {
    const {url} = req.body
    await controller.addImage(url)
    res.status(200).json({
        data: "ok"
    })
})

app.get('/api/image', (req, res) => {
    const {index} = req.query
    const img = controller.getImageDetails(index)
    res.status(200).json({
        data: img
    })
})

app.get('/api/images', (req, res) => {
    const imageNames = controller.getImages()
    res.status(200).json({
        data: imageNames
    })
})

app.use(express.static('client'))

app.listen(port, () => console.log(`Listening at http://localhost:${port}`))
