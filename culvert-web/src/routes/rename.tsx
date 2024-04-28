import React, { useEffect } from "react"

const Rename = () => {
  useEffect(() => {
    const queryString = window.location.search
    const query = new URLSearchParams(queryString)
    console.log(query)
  }, [])
  return <div>Rename</div>
}

export default Rename
