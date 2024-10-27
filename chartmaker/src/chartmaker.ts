import { createCanvas } from "@napi-rs/canvas"
import { ChartItem } from "chart.js"
import Chart from "chart.js/auto"
import ChartDataLabels from "chartjs-plugin-datalabels"
import { fontFamily } from "./fontfamily"

export const chartmaker = (
  data: { label: string; score: number }[],
): Buffer => {
  const width = data.length <= 8 ? 1000 : 125 * data.length
  const height = 600
  const chartJSNodeCanvas = createCanvas(width, height)
  const ctx = chartJSNodeCanvas.getContext("2d")
  const labels: string[] = []
  const rawData: number[] = []
  data.forEach(row => {
    labels.push(row.label)
    rawData.push(row.score)
  })

  // This works. yes.
  const chart = new Chart(ctx as unknown as ChartItem, {
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
            family: fontFamily,
          },
          formatter: Math.round,
          padding: 6,
        },
      },
    },
  })
  const b = chartJSNodeCanvas.toBuffer("image/png")
  chart.destroy()
  return b
}
