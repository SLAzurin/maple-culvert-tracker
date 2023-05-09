import { Login } from "./features/login/Login"
import "./App.css"
import { useAppSelector } from "./app/hooks"
import { selectToken } from "./features/login/loginSlice"
import { useEffect } from "react"
import fetchMembers from "./helpers/fetchMembers"

function App() {
  const token = useAppSelector(selectToken)

  useEffect(() => {
    ;(async () => {
      await fetchMembers(token)
    })()
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
