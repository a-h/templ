let src = new EventSource("/_templ/reload/events");
src.onmessage = (event) => {
	if (event && event.data === "reload") {
		window.location.reload();
	}
};
