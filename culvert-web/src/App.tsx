import { useEffect, useState } from "react"

interface MCTClaims {
  exp: string
  discord_username: string
}

function App() {
  const [claims, setClaims] = useState<MCTClaims>({
    exp: "",
    discord_username: "",
  })
  const [auth, setAuth] = useState<string>(
    localStorage.getItem("auth")
      ? JSON.parse(localStorage.getItem("auth") as string)
      : "",
  )

  useEffect(() => {
    if (auth !== "") {
      setClaims(
        JSON.parse(Buffer.from(auth.split(".")[1], "base64").toString()),
      )
    }
  }, [auth])

  return (
    <>
      Hi
    </>
  )
}

export default App
