import { Login } from "./features/login/Login"
import "./App.css"
import { useEffect, useState } from "react"
import fetchMembers from "./helpers/fetchMembers"
import {
  resetToken,
  selectClaims,
  selectToken,
} from "./features/login/loginSlice"
import { setMembers } from "./features/members/membersSlice"
import { store } from "./app/store"
import { useSelector } from "react-redux"

function App() {
  const token = useSelector(selectToken)
  const claims = useSelector(selectClaims)
  const [action, setAction] = useState("")
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
              <option value={"link_member"}>Link member's discord</option>
              <option value={"culvert_score"}>Add culvert score</option>
            </select>
          </div>
        )}
        {action === "link_member" && <div>WIP</div>}
        {action === "culvert_score" && <div>WIP</div>}
      </header>
    </div>
  )
}

export default App
