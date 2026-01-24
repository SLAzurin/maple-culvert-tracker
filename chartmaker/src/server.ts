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

const server = Bun.serve({
  routes: {
    "/chartmaker/ping": Response.json("Alive!"),
    "/chartmaker-multiple": {
      POST: async req => {
        let body: any
        try {
          body = await req.json()
        } catch (e) {}
        if (
          typeof body !== "object" ||
          !Array.isArray(body.labels) ||
          !body.labels.every((label: string) => typeof label === "string") ||
          !Array.isArray(body.dataPlots) ||
          !body.dataPlots.every(
            (plot: { characterName: string; scores: number[] }) =>
              typeof plot.characterName === "string" &&
              Array.isArray(plot.scores) &&
              plot.scores.every(score => typeof score === "number") &&
              body.labels.length === plot.scores.length,
          )
        ) {
          return Response.json("Invalid input data format", {
            status: 400,
          })
        }
        return new Response(chartmakerMultiple(body) as any, {
          headers: {
            "Content-Type": "image/png",
          },
          status: 200,
        })
      },
    },
    "/chartmaker": {
      POST: async req => {
        let body: any
        try {
          body = await req.json()
        } catch (e) {}
        if (
          !Array.isArray(body) ||
          !body.every(
            row =>
              typeof row.label === "string" && typeof row.score === "number",
          )
        ) {
          return Response.json("Invalid input data format", {
            status: 400,
          })
        }
        // check http query for yAxisStartAt0
        const url = new URL(req.url)
        const yAxisStartAt0Query = url.searchParams.get("y-axis-start-at-0")
        let yAxisStartAt0 = false
        if (yAxisStartAt0Query === "true") {
          yAxisStartAt0 = true
        }
        return new Response(chartmaker(body, { yAxisStartAt0 }) as any, {
          headers: {
            "Content-Type": "image/png",
          },
          status: 200,
        })
      },
    },
  },
})

console.log(`Chartmaker server running at ${server.url}`)
