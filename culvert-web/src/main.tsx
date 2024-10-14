import React from "react";
import ReactDOM from "react-dom/client";
import { Provider } from "react-redux";
import { persistor, store } from "./app/store";
import App from "./App";
import "./index.css";
import { PersistGate } from "redux-persist/integration/react";
import {
	createBrowserRouter,
	createRoutesFromElements,
	Route,
	RouterProvider,
} from "react-router-dom";

// Bootstrap CSS
import "bootstrap/dist/css/bootstrap.min.css";
// Bootstrap Bundle JS
import "bootstrap/dist/js/bootstrap.bundle.min";
import Rename from "./routes/rename";
import NewChar from "./routes/newchar";
import LinkDiscord from "./routes/LinkDiscord";
import EditSettings from "./routes/EditSettings";

const router = createBrowserRouter(
	createRoutesFromElements([
		<Route path="/" element={<App />} />,
		<Route path="/rename" element={<Rename />} />,
		<Route path="/newchar" element={<NewChar />} />,
		<Route path="/linkdiscord" element={<LinkDiscord />} />,
		<Route path="/edit-settings" element={<EditSettings />} />,
	]),
);

ReactDOM.createRoot(document.getElementById("root")!).render(
	<React.StrictMode>
		<Provider store={store}>
			<PersistGate loading={null} persistor={persistor}>
				<RouterProvider router={router} />
			</PersistGate>
		</Provider>
	</React.StrictMode>,
);
