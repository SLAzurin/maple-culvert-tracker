import express from "express"
import { chartmaker } from "./chartmaker"
import { chartmakerMultiple } from "./chartmaker-multiple"
import ChartDataLabels from "chartjs-plugin-datalabels"
import { Chart } from "chart.js"

Chart.register(ChartDataLabels, {
  id: "BackgroundColor",
  beforeDraw: chart => {
    const { ctx } = chart
    ctx.save()
    ctx.fillStyle = "rgba(27,27,27,255)"
    ctx.fillRect(0, 0, chart.canvas.width, chart.canvas.height)
    ctx.restore()
  },
})

const app = express()
const port = process.env.PORT || 3000

app.use(express.json())

app.get("/", (req, res) => {
  res.statusCode = 200
  res.send("Alive!")
})

app.post("/chartmaker-multiple", (req, res) => {
  res.statusCode = 200
  res.type("png")
  res.send(chartmakerMultiple(req.body))
})

app.post("/chartmaker", (req, res) => {
  res.statusCode = 200
  res.type("png")
  res.send(chartmaker(req.body))
})

app.listen(port, () => {
  console.log(`Chartmaker server listening on port ${port}`)
})
