import { useEffect, useState } from "react"

interface MCTClaims {
  exp: string
  discord_username: string
  discord_server_id: string
}

const defaultClaims: MCTClaims = {
  exp: "0",
  discord_username: "",
  discord_server_id: "",
}

interface GuildMember {
  character_name: string
  discord_username: string
}

const fetchMembers = async (auth: string): Promise<GuildMember[] | number> => {
  const claims: MCTClaims = JSON.parse(window.atob(auth.split(".")[1]))
  if (Number(claims.exp) * 1000 <= new Date().getTime()) {
    return -9 // I use -9 as a code to delete the auth
  }
  try {
    const res = await fetch(
      `${import.meta.env.VITE_BACKEND_SCHEME}://${
        import.meta.env.VITE_BACKEND_HOST
      }/api/discord/${claims.discord_server_id}/members`,
      {
        headers: {
          Authorization: `Bearer ${auth}`,
        },
      },
    )
    if (res.status !== 200) {
      return Promise.resolve(res.status)
    }
    return await res.json()
  } catch (e) {
    return Promise.resolve(-1)
  }
}

function App() {
  const [claims, setClaims] = useState<MCTClaims>(defaultClaims)
  const [members, setMembers] = useState<GuildMember[]>([])
  const [auth, setAuth] = useState<string>(
    localStorage.getItem("auth")
      ? (localStorage.getItem("auth") as string)
      : "",
  )

  useEffect(() => {
    if (auth !== "") {
      (async () => {
        const newmembers = await fetchMembers(auth)
        if (typeof newmembers === "number") {
          console.log("HTTP status not 200", newmembers)
          setAuth("")
          if (-9 === newmembers) {
            localStorage.removeItem("auth")
          }
          return
        }
        setClaims(JSON.parse(window.atob(auth.split(".")[1])))
        setMembers(newmembers as GuildMember[])
      })()
    } else {
      setClaims(defaultClaims)
    }
    console.log(auth)
  }, [auth])

  useEffect(() => {
    console.log(claims, new Date().getTime())
  }, [claims])

  const submitAction = (fn: () => void) => {
    if (auth == "") {
      return
    }

    if (new Date().getTime() >= Number(claims.exp) * 1000) {
      // expired
      setAuth("")
      setClaims(defaultClaims)
      return
    }

    fn()
  }

  return (
    <>
      <div className="main-container">
        <h4>
          Login token:{" "}
          <span>
            {claims &&
              claims.exp !== "0" &&
              "Expires " + new Date(Number(claims.exp) * 1000).toString()}
          </span>
        </h4>
        <input
          type="text"
          value={auth}
          onChange={(e) => {
            setAuth(e.target.value)
          }}
          onFocus={(e) => {
            e.target.select()
          }}
        ></input>
      </div>
      {members.map((m) => (
        <div>{m.character_name}</div>
      ))}
      <button
        onClick={() => {
          submitAction(() => {
            console.log("submitAction to be implemented")
          })
        }}
      >
        Submit
      </button>
    </>
  )
}

export default App
