import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useSelector } from "react-redux";
import { selectToken } from "../features/login/loginSlice";

const EditGlobalSettings = () => {
	const navigate = useNavigate();
	const token = useSelector(selectToken);
	const [status, setStatus] = useState("");
	const [editableValues, setEditableValues] = useState({});

	useEffect(() => {
		
	}, []);

	return (
		<div>
			<button className="btn btn-secondary" onClick={() => navigate(-1)}>
				Return to homepage
			</button>

			<form></form>
		</div>
	);
};

export default EditGlobalSettings;
