// You can use npm install to add additional packages.
// And import them in this file.
// esbuild will bundle them into the final output.

interface Data {
	msg: string;
}

function setupAttributeAlerter() {
	const alerter = document.querySelector("#attributeAlerter");
	if (!alerter) {
		return;
	}
	alerter.addEventListener("click", (_event: Event) => {
		const dataAttr = alerter?.getAttribute('alert-data') ?? '{}';
		const data: Data = JSON.parse(dataAttr);
		alert(data.msg);
	})
}

function setupScriptAlerter() {
	const alerter = document.querySelector("#scriptAlerter");
	if (!alerter) {
		return;
	}
	alerter.addEventListener("click", (_event: Event) => {
		const dataContent = document?.getElementById('scriptData')?.textContent ?? '{}';
		const data: Data = JSON.parse(dataContent);
		alert(data.msg);
	})
}

setupAttributeAlerter();
setupScriptAlerter();
