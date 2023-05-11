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
import linkDiscordMaple from "./helpers/linkDiscordMaple"
import {
  selectCharacterScores,
  selectCharacters,
  setCharacterScores,
  setCharacters,
} from "./features/characters/charactersSlice"
import fetchCharacters from "./helpers/fetchCharacters"
import fetchCharacterScores from "./helpers/fetchCharacterScores"

function App() {
  const token = useSelector(selectToken)
  const claims = useSelector(selectClaims)
  const members = useSelector(selectMembers)
  const characters = useSelector(selectCharacters)
  const characterScores = useSelector(selectCharacterScores)
  const [action, setAction] = useState("")
  const [disabledLink, setDisabledLink] = useState(false)
  const [linkCharacterName, setLinkCharacterName] = useState("")
  const [statusMessage, setStatusMessage] = useState("")
  const [successful, setSuccessful] = useState(true)

  const [selectedDiscordID, setSelectedDiscordID] = useState(
    members.length !== 0 ? members[0].discord_user_id : "",
  )

  useEffect(() => {
    if (action === "culvert_score" && Object.values(characters).length === 0) {
      console.log("action get characters")
      fetchCharacters(token).then((res) => {
        if (typeof res === "number") {
          setSuccessful(false)
          setStatusMessage("Failed with error " + res)
          return
        }
        store.dispatch(setCharacters(res))
      })
    }
  }, [action, characters, token])
  useEffect(() => {
    if (
      action === "culvert_score" &&
      Object.values(characters).length !== 0 &&
      Object.values(characterScores).length === 0
    ) {
      console.log("action get character scores")
      fetchCharacterScores(token).then((res) => {
        if (typeof res === "number") {
          setSuccessful(false)
          setStatusMessage("Failed with error " + res)
          return
        }
        store.dispatch(setCharacterScores(res))
      })
    }
  }, [action, characters, token, characterScores])
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
        {statusMessage !== "" && (
          <div className="m-5" style={{ color: successful ? "green" : "red" }}>
            {statusMessage}
          </div>
        )}
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
            <input
              type="text"
              placeholder="character name"
              value={linkCharacterName}
              onChange={(e) => {
                setLinkCharacterName(e.target.value)
              }}
            ></input>
            <button
              className="btn btn-primary"
              onClick={() => {
                setDisabledLink(true)
                linkDiscordMaple(
                  token,
                  selectedDiscordID,
                  linkCharacterName,
                  true,
                ).then((res) => {
                  setDisabledLink(false)
                  if (res.status !== 200) {
                    setStatusMessage(res.payload)
                    setSuccessful(false)
                  } else {
                    setStatusMessage("Successfully linked " + linkCharacterName)
                    setSuccessful(true)
                  }
                })
              }}
              disabled={disabledLink}
            >
              link
            </button>
            <div className="mt-5">
              <button
                className="btn btn-danger"
                onClick={() => {
                  console.log("unlinking character")
                  linkDiscordMaple(
                    token,
                    selectedDiscordID,
                    linkCharacterName,
                    false,
                  ).then((res) => {
                    setDisabledLink(false)
                    if (res.status !== 200) {
                      setStatusMessage(res.payload)
                      setSuccessful(false)
                    } else {
                      setStatusMessage(
                        "Successfully unlinked " + linkCharacterName,
                      )
                      setSuccessful(true)
                    }
                  })
                }}
                disabled={disabledLink}
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
