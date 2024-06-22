import { ChartJSNodeCanvas } from "chartjs-node-canvas"
import { ChartConfiguration, ScriptableContext } from "chart.js"
import Chart from "chart.js/auto"
import ChartDataLabels from "chartjs-plugin-datalabels"

Chart.register(ChartDataLabels)
const dataColors = [
  "#ea5545",
  "#f46a9b",
  "#ef9b20",
  "#edbf33",
  "#ede15b",
  "#bdcf32",
  "#87bc45",
  "#27aeef",
  "#b33dc6",
] as const

export const chartmakerMultiple = (data: {
  labels: string[]
  dataPlots: { characterName: string; scores: number[] }[]
}): Buffer => {
  const width = data.labels.length <= 8 ? 1000 : 125 * data.labels.length
  const height = 600
  const backgroundColour = "rgba(27,27,27,255)"
  const chartJSNodeCanvas = new ChartJSNodeCanvas({
    width,
    height,
    backgroundColour,
  })

  const lineChartConfig: ChartConfiguration<any> = {
    plugins: [ChartDataLabels],
    type: "line",
    data: {
      labels: data.labels,
      datasets: data.dataPlots.map((lineData, i) => {
        return {
          label: lineData.characterName,
          data: lineData.scores,
          datalabels: {
            align: "end",
            anchor: "end",
          },
          fill: true,
          borderColor: dataColors[i % dataColors.length],
          tension: 0.1,
          spanGaps: true,
        }
      }),
    },
    options: {
      plugins: {
        datalabels: {
          backgroundColor: function (context: any) {
            // https://chartjs-plugin-datalabels.netlify.app/guide/options.html#option-context
            return dataColors[context.datasetIndex % dataColors.length]
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
