(function() {
  let templ_reloadSrc = window.templ_reloadSrc || new EventSource("/_templ/reload/events");
  templ_reloadSrc.onmessage = (event) => {
    if (event && event.data === "reload") {
      window.location.reload();
    }
  };
  window.templ_reloadSrc = templ_reloadSrc;
  window.onbeforeunload = () => window.templ_reloadSrc.close();
})();
