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
  addNewCharacterScore,
  applyCulvertChanges,
  resetCharacterScores,
  selectCharacterScores,
  selectCharacters,
  selectMembersCharacters,
  selectEditableWeeks,
  selectSelectedWeek,
  selectUpdateCulvertScoresResult,
  setCharacterScores,
  setCharacters,
  setSelectedWeek,
  updateScoreValue,
} from "./features/characters/charactersSlice"
import fetchCharacters from "./helpers/fetchCharacters"
import fetchCharacterScores from "./helpers/fetchCharacterScores"
import { selectMembersByID } from "./features/members/membersSlice"
import renameCharacter from "./helpers/renameCharacter"

interface ImportedData {
  [key: string]: number
}

function App() {
  const token = useSelector(selectToken)
  const claims = useSelector(selectClaims)
  const members = useSelector(selectMembers)
  const membersByID = useSelector(selectMembersByID)
  const characters = useSelector(selectCharacters)
  const characterScores = useSelector(selectCharacterScores)
  const updateCulvertScoresResult = useSelector(selectUpdateCulvertScoresResult)
  const membersCharacters = useSelector(selectMembersCharacters)
  const editableWeeks = useSelector(selectEditableWeeks)
  const selectedWeek = useSelector(selectSelectedWeek)
  const [action, setAction] = useState("")
  const [searchDiscordID, setSearchDiscordID] = useState("")
  const [searchMode, setSearchMode] = useState("text")
  const [disabledLink, setDisabledLink] = useState(false)
  const [linkCharacterName, setLinkCharacterName] = useState("")
  const [statusMessage, setStatusMessage] = useState("")
  const [successful, setSuccessful] = useState(true)
  const [selectedWeekFE, setSelectedWeekFE] = useState("")
  const [selectedCharacterID, setSelectedCharacterID] = useState(0)
  const [searchCharacter, setSearchCharacter] = useState("")
  const [newCharacterName, setNewCharacterName] = useState("")
  const [importedData, setImportedData] = useState("")
  const [importedDataStatus, setImportedDataStatus] = useState("")

  const [selectedDiscordID, setSelectedDiscordID] = useState(
    members.length !== 0 ? members[0].discord_user_id : "",
  )

  useEffect(() => {
    if (selectedWeekFE !== "") {
      store.dispatch(setSelectedWeek(selectedWeekFE))
    }
  }, [selectedWeekFE])

  useEffect(() => {
    if (token !== "" && action === "culvert_score" && selectedWeek !== null) {
      fetchCharacterScores(token, selectedWeek).then((res) => {
        if (typeof res === "number") {
          setSuccessful(false)
          setStatusMessage("Failed with error " + res)
          return
        }
        store.dispatch(setCharacterScores(res))
      })
    }
  }, [selectedWeek, token, action])

  useEffect(() => {
    if (updateCulvertScoresResult !== null) {
      updateCulvertScoresResult.then((res) => {
        setDisabledLink(false)
        setSuccessful(res.status === 200)
        setStatusMessage(res.statusMessage)
        store.dispatch(resetCharacterScores())
      })
    }
  }, [updateCulvertScoresResult])

  useEffect(() => {
    if (
      (action === "culvert_score" ||
        action === "rename_character" ||
        action === "link_member") &&
      Object.values(characters).length === 0
    ) {
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
      characterScores === null
    ) {
      console.log("action get character scores")
      fetchCharacterScores(
        token,
        selectedWeek !== null ? selectedWeek : "",
      ).then((res) => {
        if (typeof res === "number") {
          setSuccessful(false)
          setStatusMessage("Failed with error " + res)
          return
        }
        store.dispatch(setCharacterScores(res))
      })
    }
  }, [action, characters, token, characterScores, selectedWeek])
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
            // Using store's dispatch to go around react hook exhaustive deps
            store.dispatch(resetToken())
          }
          return
        }
        store.dispatch(setMembers(res))
        setAction("culvert_score")
      })()
    }
  }, [token, claims])

  useEffect(() => {
    // Handle importedData onChange
    if (!importedData) {
      return
    }
    setImportedDataStatus("")
    let importedDataObj: ImportedData
    try {
      importedDataObj = JSON.parse(importedData)
    } catch (e) {
      setImportedDataStatus("Failed to import. Likely wrong syntax.")
      return
    }
    let characterMap: { [key: string]: number } = {}
    for (const [id, char] of Object.entries(characters)) {
      characterMap[char] = Number(id)
    }
    let importedScores: { [key: number]: number } = {}
    let scoreErrors: { [key: string]: number } = {}
    for (const [charName, score] of Object.entries(importedDataObj)) {
      if (characterMap[charName]) {
        importedScores[characterMap[charName]] = score
      } else {
        scoreErrors[charName] = score
      }
    }
    for (const [id, score] of Object.entries(importedScores)) {
      store.dispatch(addNewCharacterScore(Number(id)))
      store.dispatch(updateScoreValue({ character_id: Number(id), score }))
    }
    if (Object.keys(scoreErrors).length === 0) {
      setImportedDataStatus("Successfully imported all characters")
    } else {
      setImportedDataStatus(
        "Imported partially, errors with these chars/scores\n" +
          JSON.stringify(scoreErrors, null, 2),
      )
    }
    setImportedData("")
  }, [importedData, characters])

  return (
    <div className="App">
      <header className="App-header">
        <Login />
        {statusMessage !== "" && (
          <div className="m-5" style={{ color: successful ? "green" : "red" }}>
            {statusMessage}
          </div>
        )}
        {/* {claims.exp !== "0" && (
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
              <option value={"rename_character"}>Rename character</option>
            </select>
          </div>
        )} */}
        {/* {action === "link_member" && members.length !== 0 && (
          <div>
            <div>
              Search discord member by
              <div>
                <input
                  id="search-mode-text"
                  type="radio"
                  name="search-mode"
                  value="text"
                  onChange={(e) => {
                    setSearchMode(e.target.value)
                  }}
                  checked={"text" === searchMode}
                />
                <label htmlFor="search-mode-text">Discord username/ID</label>
                <br />
                <input
                  id="search-mode-dropdown"
                  type="radio"
                  name="search-mode"
                  value="dropdown"
                  onChange={(e) => {
                    setSearchMode(e.target.value)
                  }}
                  checked={"dropdown" === searchMode}
                />
                <label htmlFor="search-mode-dropdown">
                  Dropdown of all members
                </label>
              </div>
            </div>
            {searchMode === "text" && (
              <div>
                {selectedDiscordID !== "" && (
                  <div>Selected {membersByID[selectedDiscordID]}</div>
                )}
                <input
                  type="text"
                  placeholder="Discord username / ID"
                  value={searchDiscordID}
                  onChange={(e) => {
                    setSearchDiscordID(e.target.value)
                  }}
                />
                {searchDiscordID !== "" && (
                  <div>
                    {members
                      .filter((m) => {
                        return (
                          (m.discord_nickname || "")
                            .toLowerCase()
                            .includes(searchDiscordID.toLowerCase()) ||
                          m.discord_username
                            .toLowerCase()
                            .includes(searchDiscordID.toLowerCase()) ||
                          m.discord_user_id
                            .toLowerCase()
                            .includes(searchDiscordID.toLowerCase()) ||
                          m.discord_global_name
                            .toLowerCase()
                            .includes(searchDiscordID.toLowerCase())
                        )
                      })
                      .map((m) => (
                        <button
                          key={"link_member-search-" + m.discord_user_id}
                          className="btn btn-success"
                          onClick={() => {
                            setSelectedDiscordID(m.discord_user_id)
                          }}
                        >
                          {m.discord_nickname ||
                            m.discord_global_name ||
                            m.discord_username}
                        </button>
                      ))}
                  </div>
                )}
              </div>
            )}
            {searchMode === "dropdown" && (
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
                    {member.discord_nickname ||
                      member.discord_global_name ||
                      member.discord_username}
                  </option>
                ))}
              </select>
            )}
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
                    setSuccessful(false)
                    setStatusMessage(res.payload)
                  } else {
                    setSuccessful(true)
                    setStatusMessage("Successfully linked " + linkCharacterName)
                    store.dispatch(setCharacters([]))
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
                      store.dispatch(setCharacters([]))
                    }
                  })
                }}
                disabled={disabledLink}
              >
                unlink
              </button>
            </div>
            {selectedDiscordID !== "" && (
              <div>
                Currently linked characters:
                {membersCharacters[selectedDiscordID] &&
                  membersCharacters[selectedDiscordID].map((cid) => (
                    <div key={"linked-char-display-" + cid}>
                      {characters[cid]}
                    </div>
                  ))}
              </div>
            )}
          </div>
        )} */}
        {action === "culvert_score" && (
          <div>
            {editableWeeks !== null && (
              <div style={{ display: "flex", flexDirection: "column" }}>
                <button
                  className="btn btn-link"
                  onClick={() => {
                    navigator.clipboard.writeText(
                      JSON.stringify(Object.values(characters), null, 4),
                    )
                    alert("copied")
                  }}
                >
                  Copy members to clipboard
                </button>
                <textarea
                  style={{ resize: "none" }}
                  value={importedData}
                  rows={3}
                  placeholder="Select date first, then
Paste data here to quickly set values.
Don't forget to submit"
                  onChange={(e) => {
                    setImportedData(e.target.value)
                  }}
                ></textarea>
                {importedDataStatus !== "" && <p>{importedDataStatus}</p>}
                <select
                  onChange={(e) => {
                    setSelectedWeekFE(e.target.value)
                  }}
                >
                  {editableWeeks.map((d) => (
                    <option key={`editable-weeks-${d}`} value={d}>
                      {d}
                    </option>
                  ))}
                </select>
              </div>
            )}
            <table>
              <thead>
                <tr>
                  <th>Discord user</th>
                  <th>Character name</th>
                  <th>Last week</th>
                  <th>This week</th>
                </tr>
              </thead>
              <tbody>
                {Object.entries(characterScores || {})
                  .sort(([charID1], [charID2]) => {
                    return characters[Number(charID1)] >=
                      characters[Number(charID2)]
                      ? 1
                      : -1
                  })
                  .map(([charID, scores], i) => {
                    if (!characters[Number(charID)]) return null
                    return (
                      <tr className="" key={"scores-" + i}>
                        <td>
                          <span>
                            {membersCharacters &&
                              Object.entries(membersCharacters).map(
                                ([discordID, charIDs], i) => {
                                  if (
                                    charIDs.includes(Number(charID)) &&
                                    membersByID[discordID]
                                  ) {
                                    const member = members.find((member) => {
                                      return (
                                        member.discord_user_id === discordID
                                      )
                                    })
                                    return (
                                      <button
                                        key={"discord_name-button-" + i}
                                        className="btn btn-link"
                                        onClick={() => {
                                          console.log(
                                            "Button clicked for discord_name",
                                          )
                                        }}
                                      >
                                        {member?.discord_nickname ||
                                          member?.discord_global_name ||
                                          member?.discord_username ||
                                          membersByID[discordID]}
                                      </button>
                                    )
                                  }
                                  return null
                                },
                              )}
                          </span>
                        </td>
                        <td>
                          <span>{characters[Number(charID)] || charID}</span>
                        </td>
                        <td>
                          <input
                            placeholder={scores.prev?.toString()}
                            disabled={true}
                          />
                        </td>
                        <td>
                          <input
                            onChange={(e) => {
                              const n = Number(e.target.value)
                              if (!Number.isNaN(n)) {
                                store.dispatch(
                                  updateScoreValue({
                                    score: n,
                                    character_id: Number(charID),
                                  }),
                                )
                              }
                            }}
                            value={scores.current || ""}
                          />
                        </td>
                      </tr>
                    )
                  })}
              </tbody>
            </table>
            <div>
              Add character score
              <select
                value=""
                onChange={(e) => {
                  store.dispatch(addNewCharacterScore(Number(e.target.value)))
                }}
              >
                <option value={""}></option>
                {Object.keys(characters)
                  .filter(
                    (charID) =>
                      typeof (characterScores || {})[Number(charID)] ===
                      "undefined",
                  )
                  .sort((aKey, bKey) => {
                    return characters[Number(aKey)] > characters[Number(bKey)]
                      ? 1
                      : -1
                  })
                  .map((charID) => (
                    <option value={charID} key={"addnewcharacter-" + charID}>
                      {characters[Number(charID)]}
                    </option>
                  ))}
              </select>
            </div>
            <button
              disabled={disabledLink}
              className="btn btn-primary"
              onClick={() => {
                setImportedDataStatus("")
                setDisabledLink(true)
                console.log("apply changes for culvert scores")
                store.dispatch(applyCulvertChanges(token))
              }}
            >
              Submit
            </button>
          </div>
        )}
        {action === "rename_character" && (
          <div>
            <div>
              {selectedCharacterID !== 0 && (
                <div>Selected: {characters[selectedCharacterID]}</div>
              )}
              <input
                type="text"
                placeholder="character name"
                value={searchCharacter}
                onChange={(e) => {
                  setSearchCharacter(e.target.value)
                }}
              />
              {searchCharacter !== "" && (
                <div>
                  {Object.keys(characters)
                    .filter((m) => {
                      return (
                        characters[Number(m)]
                          .toLowerCase()
                          .includes(searchCharacter.toLowerCase()) ||
                        characters[Number(m)]
                          .toLowerCase()
                          .includes(searchCharacter.toLowerCase()) ||
                        characters[Number(m)]
                          .toLowerCase()
                          .includes(searchCharacter.toLowerCase())
                      )
                    })
                    .map((m) => (
                      <button
                        key={"rename_character-select-character-" + m}
                        className="btn btn-success"
                        onClick={() => {
                          setSelectedCharacterID(Number(m))
                        }}
                      >
                        {characters[Number(m)]}
                      </button>
                    ))}
                </div>
              )}
            </div>
            <input
              onChange={(e) => {
                setNewCharacterName(e.target.value)
              }}
              value={newCharacterName}
              placeholder="new name"
            />
            <br />
            <button
              className="btn btn-danger"
              disabled={disabledLink}
              onClick={() => {
                setDisabledLink(true)
                renameCharacter(token, {
                  character_id: selectedCharacterID,
                  new_name: newCharacterName,
                }).then((res) => {
                  setDisabledLink(false)
                  if (res.status !== 200) {
                    setSuccessful(false)
                    setStatusMessage(res.payload)
                  } else {
                    setSuccessful(true)
                    setStatusMessage(
                      "Successfully renamed to " + newCharacterName,
                    )
                    store.dispatch(setCharacters([]))
                  }
                })
              }}
            >
              Rename
            </button>
          </div>
        )}
      </header>
    </div>
  )
}

export default App
