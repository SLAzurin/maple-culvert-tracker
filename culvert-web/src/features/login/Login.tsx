import { useAppSelector, useAppDispatch } from "../../app/hooks"
import { selectClaims, setToken, selectToken } from "./loginSlice"

export function Login() {
  const token = useAppSelector(selectToken)
  const claims = useAppSelector(selectClaims)
  const dispatch = useAppDispatch()

  return (
    <div>
      {claims &&
        claims.exp !== "0" &&
        "Expires " + new Date(Number(claims.exp) * 1000).toString()}
      <div>
        Login token:{" "}
        <input
          type="text"
          onChange={(e) => {
            dispatch(setToken(e.target.value))
          }}
          value={token}
        />
      </div>
    </div>
  )
}
