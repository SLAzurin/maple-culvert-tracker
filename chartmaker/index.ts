import fs from "fs"
import { ChartJSNodeCanvas } from "chartjs-node-canvas"
import { ChartConfiguration } from "chart.js"
import Chart from "chart.js/auto"
import ChartDataLabels from "chartjs-plugin-datalabels"
import sharp from "sharp"

Chart.register(ChartDataLabels)

const data: { label: string; score: number }[] = JSON.parse(
  process.env.DATA
    ? process.env.DATA
    : fs.readFileSync("sample.json").toString(),
)

const width = 1000
const height = 600
const backgroundColour = "rgba(0, 0, 0, 0.1)"
const chartJSNodeCanvas = new ChartJSNodeCanvas({
  width,
  height,
  backgroundColour,
})

const labels: string[] = []
const rawData: number[] = []
data.forEach(row => {
  labels.push(row.label)
  rawData.push(row.score)
})

const lineChartConfig: ChartConfiguration<any> = {
  plugins: [ChartDataLabels],
  type: "line",
  data: {
    labels: labels,
    datasets: [
      {
        label: "Culvert score by week",
        data: rawData,
        datalabels: {
          align: "end",
          anchor: "end",
        },
        fill: false,
        borderColor: "rgba(54, 162, 235, 1)",
        tension: 0.1,
        spanGaps: true,
      },
    ],
  },
  options: {
    plugins: {
      datalabels: {
        backgroundColor: function () {
          return "rgba(54, 162, 235, 1)"
        },
        borderRadius: 4,
        color: "white",
        font: {
          weight: "bold",
        },
        formatter: Math.round,
        padding: 6,
      },
    },
  },
}

sharp("bg.png")
  .composite([
    {
      input: chartJSNodeCanvas.renderToBufferSync(lineChartConfig),
    },
  ])
  .toFile("img.png")

// fs.writeFileSync(
//   "img.png",
//   chartJSNodeCanvas.renderToBufferSync(lineChartConfig),
// )
