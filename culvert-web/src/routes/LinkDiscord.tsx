import React, { useEffect } from "react"

const LinkDiscord = () => {
  useEffect(() => {
    const queryString = window.location.search
    const query = new URLSearchParams(queryString)
    console.log(query)
  }, [])
  return <div>LinkDiscord</div>
}

export default LinkDiscord
