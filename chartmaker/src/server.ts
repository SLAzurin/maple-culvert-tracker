import express from "express"
import * as child_process from "child_process"
const app = express()
const port = 3005

app.use(express.json())

app.get("/", (req, res) => {
  res.statusCode = 200
  res.send("Alive!")
})

app.post("/chartmaker", (req, res) => {
  res.statusCode = 200
  const buffer = child_process.spawnSync(
    "/usr/local/bin/node", // absolute path in container
    // process.env.HOME + "/.nvm/versions/node/v18.16.0/bin/node",
    ["dist/chartmaker.js"],
    {
      env: {
        DATA: JSON.stringify(req.body),
      },
    },
  )
  res.type("png")
  res.send(Buffer.from(buffer.stdout.toString(), "base64"))
})

app.listen(port, () => {
  console.log(`Chartmaker server listening on port ${port}`)
})
