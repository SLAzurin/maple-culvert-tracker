import fs from "fs"
import { ChartJSNodeCanvas } from "chartjs-node-canvas"
import { ChartConfiguration } from "chart.js"

const data: { label: string; score: number }[] = JSON.parse(
  process.env.DATA
    ? process.env.DATA
    : fs.readFileSync("sample.json").toString(),
)

const width = 1000
const height = 600
const backgroundColour = 'rgba(0, 0, 0, 0.1)'
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
  type: "line",
  data: {
    labels: labels,
    datasets: [
      {
        label: "Score by week",
        data: rawData,
        fill: false,
        borderColor: "rgb(75, 192, 192)",
        tension: 0.1,
      },
    ],
  },
  options: {
    backgroundColour: "rgb(230,230,230)",
  },
}

;(async () => {
  const image = await chartJSNodeCanvas.renderToDataURL(lineChartConfig)
  console.log(image)
})()
