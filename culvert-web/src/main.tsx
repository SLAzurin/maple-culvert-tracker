import React from "react"
import ReactDOM from "react-dom/client"
import { Provider } from "react-redux"
import { persistor, store } from "./app/store"
import App from "./App"
import "./index.css"
import { PersistGate } from "redux-persist/integration/react"

// Bootstrap CSS
import "bootstrap/dist/css/bootstrap.min.css"
// Bootstrap Bundle JS
import "bootstrap/dist/js/bootstrap.bundle.min"

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <Provider store={store}>
      <PersistGate loading={null} persistor={persistor}>
        <App />
      </PersistGate>
    </Provider>
  </React.StrictMode>,
)
