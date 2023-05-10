import { Login } from "./features/login/Login"
import "./App.css"
import { useEffect, useState } from "react"
import fetchMembers from "./helpers/fetchMembers"
import {
  resetToken,
  selectClaims,
  selectToken,
} from "./features/login/loginSlice"
import { selectMembers, setMembers } from "./features/members/membersSlice"
import { store } from "./app/store"
import { useSelector } from "react-redux"

function App() {
  const token = useSelector(selectToken)
  const claims = useSelector(selectClaims)
  const members = useSelector(selectMembers)
  const [action, setAction] = useState("")

  const [selectedDiscordID, setSelectedDiscordID] = useState(
    members.length !== 0 ? members[0].discord_user_id : "",
  )
  useEffect(() => {
    // claims expired
    if (
      claims.exp !== "0" &&
      Number(claims.exp) * 1000 < new Date().getTime()
    ) {
      alert("Expired login token")
      store.dispatch(resetToken())
      return
    }
    // if new token was entered
    if (token !== "") {
      ;(async () => {
        console.log("fetching members")
        const res = await fetchMembers(token)
        if (typeof res === "number") {
          console.log("failed to get members", res)
          if (res === 401) {
            alert("Expired login token")
            // Using store's dispatch to go around react hook exhaustive deps
            store.dispatch(resetToken())
          }
          return
        }
        store.dispatch(setMembers(res))
      })()
    }
  }, [token, claims])
  return (
    <div className="App">
      <header className="App-header">
        <Login />
        {claims.exp !== "0" && (
          <div className="m-5">
            What would you like to do?
            <select
              onChange={(e) => {
                setAction(e.target.value)
              }}
              value={action}
            >
              <option value={""}></option>
              <option value={"link_member"}>
                Link discord user with maple character
              </option>
              <option value={"culvert_score"}>Add culvert score</option>
            </select>
          </div>
        )}
        {action === "link_member" && members.length !== 0 && (
          <div>
            <select
              value={selectedDiscordID}
              onChange={(e) => {
                setSelectedDiscordID(e.target.value)
              }}
            >
              {members.map((member) => (
                <option
                  key={member.discord_user_id}
                  value={member.discord_user_id}
                >
                  {member.discord_username}
                </option>
              ))}
            </select>
            <input type="text" placeholder="character name"></input>
            <button
              className="btn btn-primary"
              onClick={() => {
                console.log("Clicked link character to discord id")
              }}
            >
              link
            </button>
            <div className="mt-5">
              <button
                className="btn btn-danger"
                onClick={() => {
                  console.log("unlinking character")
                }}
              >
                unlink
              </button>
            </div>
          </div>
        )}
        {action === "culvert_score" && <div>WIP</div>}
      </header>
    </div>
  )
}

export default App
