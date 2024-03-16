let templ_reloadSrc = new EventSource("/_templ/reload/events");
templ_reloadSrc.onmessage = (event) => {
	if (event && event.data === "reload") {
		window.location.reload();
	}
};
