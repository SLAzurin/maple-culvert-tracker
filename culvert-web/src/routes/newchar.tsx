import { useState } from "react"
import { useNavigate } from "react-router-dom"
import { useSelector } from "react-redux"
import { resetInitialStateCharacters } from "../features/characters/charactersSlice"
import { selectToken } from "../features/login/loginSlice"
import { store } from "../app/store"
import linkDiscordMaple from "../helpers/linkDiscordMaple"

const NewChar = () => {
  const navigate = useNavigate()
  const token = useSelector(selectToken)
  const [status, setStatus] = useState("")
  const [characterName, setCharacterName] = useState("")

  return (
    <div>
      <h1>Track new character</h1>
      {status !== "" && <h2>Status: {status}</h2>}
      <input
        value={characterName}
        placeholder="Character Name"
        onChange={(e) => {
          setCharacterName(e.target.value)
        }}
      ></input>
      <button
        className="btn btn-primary"
        onClick={async () => {
          linkDiscordMaple(token, "2", characterName, true).then((res) => {
            if (res.status !== 200) {
              setStatus(`Error: ${res.status} ${res.payload}`)
              return
            }
            store.dispatch(resetInitialStateCharacters())
            navigate("/")
          })
        }}
      >
        Submit
      </button>
    </div>
  )
}

export default NewChar
