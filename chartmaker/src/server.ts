#!/usr/bin/env node
import express from "express"
import { chartmaker } from "./chartmaker"

const app = express()
const port = process.env.PORT || 3000

app.use(express.json())

app.get("/", (req, res) => {
  res.statusCode = 200
  res.send("Alive!")
})

app.post("/chartmaker", async (req, res) => {
  res.statusCode = 200
  let buffer = ""
  if (!req.body) {
    res.statusCode = 400
    return res.send("")
  }
  try {
    buffer = await chartmaker(JSON.stringify(req.body))
  } catch (e: any) {
    res.statusCode = 400
    return res.send("")
  }
  res.type("png")
  res.send(Buffer.from(buffer, "base64"))
})

app.listen(port, () => {
  console.log(`Chartmaker server listening on port ${port}`)
})
