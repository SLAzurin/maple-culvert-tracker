import * as dotenv from "dotenv"
dotenv.config()
import { Client, GatewayIntentBits } from "discord.js"

const client = new Client({ intents: [GatewayIntentBits.Guilds] })

client.on("ready", () => {
  console.log(`Logged in as ${client.user?.tag}!`)
})

// init all commands
client.on("interactionCreate", async (interaction) => {
  if (!interaction.isChatInputCommand()) return

  if (interaction.commandName === "ping") {
    await interaction.reply("Pong!")
  }
})

await client.login(process.env.DISCORD_TOKEN)
