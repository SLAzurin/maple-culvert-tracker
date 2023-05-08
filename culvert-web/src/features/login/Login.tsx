import { useState } from "react"
import { useAppSelector, useAppDispatch } from "../../app/hooks"
import { setToken, selectToken } from "./loginSlice"

export function Login() {
  const token = useAppSelector(selectToken)
  const dispatch = useAppDispatch()

  return (
    <div>
      Login token:{" "}
      <input
        type="text"
        onChange={(e) => {
          dispatch(setToken(e.target.value))
        }}
      />
    </div>
  )
}
