# Datastar

[Datastar](https://data-star.dev) is another hypermedia framework similar to [HTMX](htmx) can be used to selectively replace content within a web page. It does this by combining fine grain reactive signals with SSE. It's geared primarily to real time applications where you'd normally reach for a SPA framework such as React/Vue/Svelte. If you are combining HTMX with a reactive frames (such as AlpineJS or hyperscript) then you may find Datastar a better fit.

## Usage

Using Datastar requires:

- Installation of the Datastar client-side library.
- Modifying the HTML markup to instruct the library to perform partial screen updates.

## Installation

Datastar comes out of the box with templ components to speed up development. You can use `@datastar.ScriptCDNLatest()` or `ScriptCDNVersion(version string)` to include the latest version of the Datastar library in your HTML.

:::info
Advanced Datastar installation and usage help is covered in the user guide at https://data-star.dev.
:::

## Datastar examples using Templ

Datastar website itself is built using Datastar and Templ. Look at any of the examples in the [source code](https://github.com/delaneyj/datastar/tree/main/backends/go/site) for the Datastar website to see how it's done! This specific example is [here](http://data-star.dev/examples/templ_counter)

## Counter Example

We are going to modify the counter example [previously discussed](example-counter-application) to use Datastar.

### Frontend

First, define a HTML some with two buttons. One to update a global state, and one for a per-user state.

```templ title="components.templ"
package site

import (
	"github.com/delaneyj/datastar"
)

type TemplCounterStore struct {
	Global uint32 `json:"global"`
	User   uint32 `json:"user"`
}

templ templCounterExampleButtons() {
	<div>
		<button
			data-on-click="$$post('/examples/templ_counter/increment/global')"
		>
			Increment Global
		</button>
		<button
			data-on-click="$$post('/examples/templ_counter/increment/user')"
		>
			Increment User
		</button>
	</div>
}

templ templCounterExampleCounts() {
	<div>
		<div>
			<div>Global</div>
			<div data-text="$global"></div>
		</div>
		<div>
			<div>User</div>
			<div data-text="$user"></div>
		</div>
	</div>
}

templ templCounterExampleInitialContents(store TemplCounterStore) {
	<div
		id="container"
		data-store={ datastar.MustJSONString(store) }
	>
		@templCounterExampleButtons()
		@templCounterExampleCounts()
	</div>
}

```

:::tip
Note that Datastar doesn't promote the use of forms. This is because they are ill suited to nested reactive contents. Instead it will send all[^1] reactive state (as JSON) to the server on every request. This means far less bookkeeping and more predictable state management.
:::

:::note
`data-store` is a special attribute that Datastar uses to store the initial state of the page. The contents will turn into signals that can be used to update the page.

`data-on-click="$$post('/examples/templ_counter/increment/global')"` is an attribute expression that says "when this element is clicked, send a POST request to the server to the specified URL". The `$$post` is an action that is a sandboxed function that knows about things like signals.

`data-text="$global"` is an attribute expression that says "replace the contents of this element with the value of the `global` signal in the store". This is a reactive signal that will update the page when the value changes, which we'll see in a moment.
:::

### Backend

The backend is pretty straightforward. Notice there is a bit of setup for the SSE handling but Datastar comes with helpers built in to make this easy.

```go title="examples_templ_counter.go"
package site

import (
	"net/http"
	"sync/atomic"

	"github.com/Jeffail/gabs/v2"
	"github.com/delaneyj/datastar"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
)

func setupExamplesTemplCounter(examplesRouter chi.Router, sessionStore sessions.Store) error {

	var globalCounter atomic.Uint32
	const (
		sessionKey = "templ_counter"
		countKey   = "count"
	)

	userVal := func(r *http.Request) (uint32, *sessions.Session, error) {
		sess, err := sessionStore.Get(r, sessionKey)
		if err != nil {
			return 0, nil, err
		}

		val, ok := sess.Values[countKey].(uint32)
		if !ok {
			val = 0
		}
		return val, sess, nil
	}

	examplesRouter.Get("/templ_counter/data", func(w http.ResponseWriter, r *http.Request) {
		userVal, _, err := userVal(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		store := TemplCounterStore{
			Global: globalCounter.Load(),
			User:   userVal,
		}

		sse := datastar.NewSSE(w, r)
		datastar.RenderFragmentTempl(sse, templCounterExampleInitialContents(store))
	})

	updateGlobal := func(store *gabs.Container) {
		store.Set(globalCounter.Add(1), "global")
	}

	examplesRouter.Route("/templ_counter/increment", func(incrementRouter chi.Router) {
		incrementRouter.Post("/global", func(w http.ResponseWriter, r *http.Request) {
			update := gabs.New()
			updateGlobal(update)

			sse := datastar.NewSSE(w, r)
			datastar.PatchStore(sse, update)
		})

		incrementRouter.Post("/user", func(w http.ResponseWriter, r *http.Request) {
			val, sess, err := userVal(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			val++
			sess.Values[countKey] = val
			if err := sess.Save(r, w); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}

			update := gabs.New()
			updateGlobal(update)
			update.Set(val, "user")

			sse := datastar.NewSSE(w, r)
			datastar.PatchStore(sse, update)
		})
	})

	return nil
}
```

We are using the `atomic.Uint32` type to store the global state. The `userVal` function is a helper that retrieves the user's session state. The `updateGlobal` function increments the global state.

:::note
In this example, the global state is stored in RAM, and will be lost when the web server reboots. To support load-balanced web servers, and stateless function deployments, you might consider storing the state in a data store such as [NATS KV](https://docs.nats.io/using-nats/developer/develop_jetstream/kv).
:::

### Per-user session state

In a HTTP application, per-user state information is partitioned by a HTTP cookie. Cookies that identify a user while they're using a site are known as "session cookies". When the HTTP handler receives a request, it can read the session ID of the user from the cookie and retrieve any required state.

### Signal only patching

Since this is a very simple example the page's elements aren't changing dynamically we can use the `datastar.PatchStore` function to send only the signals that have changed. This is a more efficient way to update the page without even needing to send HTML fragments!

:::tip
Datastar will merge updates to the store similar to a JSON merge patch. This means you can do dynamic partial updates to the store and the page will update accordingly. [Gabs](https://pkg.go.dev/github.com/Jeffail/gabs/v2#section-readme) is a great library used here to make dynamic JSON in Go easy.

[^1]: You can control what get's sent to the server by prefixing local signals with `_`. This will prevent them from being sent to the server on every request.
