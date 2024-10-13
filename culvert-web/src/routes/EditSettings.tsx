import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useSelector } from "react-redux";
import { selectToken } from "../features/login/loginSlice";
import fetchEditableSettings from "../helpers/fetchEditableSettings";

const EditSettings = () => {
	const navigate = useNavigate();
	const token = useSelector(selectToken);
	const [status, setStatus] = useState("");
	const [editableValues, setEditableValues] = useState({} as any);

	useEffect(() => {
		try {
			fetchEditableSettings(token).then((res) => {
				setEditableValues(res);
			});
		} catch {
			alert("Failed to fetch editable settings");
		}
	}, []);

	return (
		<div>
			<button className="btn btn-secondary" onClick={() => navigate(-1)}>
				Return to homepage
			</button>

			<form>
				{Object.keys(editableValues).map((key) => (
					<div key={key}>
						<label htmlFor={key}>
							{editableValues[key]?.human_readable_description?.name}
						</label>
						<input
							id={key}
							type="text"
							defaultValue={editableValues[key].value}
						/>
					</div>
				))}
			</form>
		</div>
	);
};

export default EditSettings;
