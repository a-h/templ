// You can use npm install to add additional packages.
// And import them in this file.
// esbuild will bundle them into the final output.

interface Data {
	msg: string;
	value: number;
}

function setupAlerter() {
	const alerter = document.querySelector("#alerter");
	if (!alerter) {
		return;
	}
	alerter.addEventListener("click", (_event: Event) => {
		const dataAttr = alerter?.getAttribute('alert-data') ?? '{}';
		const data: Data = JSON.parse(dataAttr);
		alert(data.msg);
		alert(`The meaning of life etc. is ${data.value}`);
	})
}

setupAlerter();
