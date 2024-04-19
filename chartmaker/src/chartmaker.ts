import fs from "fs"
import { ChartJSNodeCanvas } from "chartjs-node-canvas"
import { ChartConfiguration } from "chart.js"
import Chart from "chart.js/auto"
import ChartDataLabels from "chartjs-plugin-datalabels"

Chart.register(ChartDataLabels)

export const chartmaker = (
  data: { label: string; score: number }[],
): Buffer => {
  const width = data.length <= 8 ? 1000 : 125 * data.length
  const height = 600
  const backgroundColour = "rgba(27,27,27,255)"
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
  return chartJSNodeCanvas.renderToBufferSync(lineChartConfig)
}
