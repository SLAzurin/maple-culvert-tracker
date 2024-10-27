import express from "express"
import { chartmaker } from "./chartmaker"
import { chartmakerMultiple } from "./chartmaker-multiple"
import ChartDataLabels from "chartjs-plugin-datalabels"
import { Chart } from "chart.js"
import { fontFamily } from "./fontfamily"

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

Chart.defaults.font.family = fontFamily

const app = express()
const port = process.env.PORT || 3000

app.use(express.json())

app.get("/chartmaker/ping", (req, res) => {
  res.statusCode = 200
  res.type("json")
  res.send('"Alive!"')
})

app.post("/chartmaker-multiple", (req, res) => {
  if (
    typeof req.body !== "object" ||
    !Array.isArray(req.body.labels) ||
    !req.body.labels.every((label: string) => typeof label === "string") ||
    !Array.isArray(req.body.dataPlots) ||
    !req.body.dataPlots.every(
      (plot: { characterName: string; scores: number[] }) =>
        typeof plot.characterName === "string" &&
        Array.isArray(plot.scores) &&
        plot.scores.every(score => typeof score === "number") &&
        req.body.labels.length === plot.scores.length,
    )
  ) {
    res.statusCode = 400
    res.type("json")
    res.send('"Invalid input data format"')
    return
  }
  res.statusCode = 200
  res.type("png")
  res.send(chartmakerMultiple(req.body))
})

app.post("/chartmaker", (req, res) => {
  if (
    !Array.isArray(req.body) ||
    !req.body.every(
      row => typeof row.label === "string" && typeof row.score === "number",
    )
  ) {
    res.statusCode = 400
    res.type("json")
    res.send('"Invalid input data format"')
    return
  }
  res.statusCode = 200
  res.type("png")
  res.send(chartmaker(req.body))
})

app.listen(port, () => {
  console.log(`Chartmaker server listening on port ${port}`)
})
