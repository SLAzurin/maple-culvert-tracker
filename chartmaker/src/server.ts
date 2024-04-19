import express from "express"
import * as child_process from "child_process"
import { chartmaker } from "./chartmaker"
const app = express()
const port = process.env.PORT || 3000

app.use(express.json())

app.get("/", (req, res) => {
  res.statusCode = 200
  res.send("Alive!")
})

app.post("/chartmaker", (req, res) => {
  const buffer = chartmaker(req.body)
  res.statusCode = 200
  res.type("png")
  res.send(buffer)
})

app.listen(port, () => {
  console.log(`Chartmaker server listening on port ${port}`)
})
