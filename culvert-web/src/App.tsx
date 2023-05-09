import { Login } from "./features/login/Login"
import "./App.css"
import { useEffect } from "react"
import fetchMembers from "./helpers/fetchMembers"
import { resetToken, selectToken } from "./features/login/loginSlice"
import { store } from "./app/store"
import { useSelector } from "react-redux"

function App() {
  const token = useSelector(selectToken)
  useEffect(() => {
    if (token !== "") {
      ;(async () => {
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
      })()
    }
  }, [token])
  return (
    <div className="App">
      <header className="App-header">
        <Login />
      </header>
    </div>
  )
}

export default App
