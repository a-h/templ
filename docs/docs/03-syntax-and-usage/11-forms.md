# Forms and validation

To pass data from the client to the server without using JavaScript, you can use HTML forms to POST data.

templ can be used to create forms that submit data to the server. Depending on the design of your app, you can collect data from the form using JavaScript and submit it to an API from the frontend, or use a HTTP form submission to send the data to the server.

## Hypermedia approach

templ isn't a framework, you're free to choose how you want to build your applications, but a common approach is to create a handler for each route, and then use templates to render the form and display validation errors.

In Go, the `net/http` package in the standard library provides a way to handle form submissions, and Gorilla `schema` can decode form data into Go structs. See https://github.com/gorilla/schema

:::tip
The [Hypermedia Systems](https://hypermedia.systems/) book covers the main concepts of building web applications, without covering specific implementations. If you're new to web development, or have only ever used JavaScript frameworks, it may be worth reading the book to understand the approach.
:::

### Create a View Model

This view model should contain any data that is used by the form, including field values and any other state.

```go
type Model struct {
  Initial          bool
  SubmitButtonText string

  Name  string
  Email string
  Error string
}
```

The model can also include methods for validation, which will be used to check the data before saving it to the database.

```go
func (m *Model) ValidateName() (msgs []string) {
  if m.Initial {
    return
  }
  if m.Name == "" {
    msgs = append(msgs, "Name is required")
  }
  return msgs
}

func (m *Model) NameHasError() bool {
  return len(m.ValidateName()) > 0
}

// More validation methods...

func (m *Model) Validate() (msgs []string) {
  if m.Initial {
    return
  }
  msgs = append(msgs, m.ValidateName()...)
  msgs = append(msgs, m.ValidateEmail()...)
  return msgs
}
```

### Create a form template

The form should contain input fields for each piece of data in the model.

In the example code, the `name` and `email` input fields are populated with the values from the model.

Later, we will use the Gorilla `schema` package to populate Go struct fields automatically from the form data when the form is submitted.

If a field value is invalid, the `has-error` class is added to the form group using the `templ.KV` function.

To protect your forms from cross-site request forgery (CSRF) attacks, use the [`gorilla/csrf`](https://github.com/gorilla/csrf) middleware to generate and validate CSRF tokens.

```go
csrfKey := mustGenerateCSRFKey()
csrfMiddleware := csrf.Protect(csrfKey, csrf.TrustedOrigins([]string{"localhost:8080"}), csrf.FieldName("_csrf"))
```

In your form templates, include a hidden CSRF token field using a shared component:

```templ
<input type="hidden" name="_csrf" value={ ctx.Value("gorilla.csrf.Token").(string) }/>
```

This ensures all POST requests include a valid CSRF token.
```templ
templ View(m Model) {
  <h1>Add Contact</h1>
  <ul>
    <li><a href="/contacts" hx-boost="true">Back to Contacts</a></li>
  </ul>
  <form id="form" method="post" hx-boost="true">
    @csrf.CSRF()
    <div id="name-group" class={ "form-group", templ.KV("has-error", m.NameHasError()) }>
      <label for="name">Name</label>
      <input type="text" id="name" name="name" class="form-control" placeholder="Name" value={ m.Name }/>
    </div>
    <div id="email-group" class={ "form-group", templ.KV("has-error", m.EmailHasError()) }>
      <label for="email">Email</label>
      <input type="email" id="email" name="email" class="form-control" placeholder="Email" value={ m.Email }/>
    </div>
    <div id="validation">
      if m.Error != "" {
        <p class="error">{ m.Error }</p>
      }
      if msgs := m.Validate(); len(msgs) > 0 {
        @ValidationMessages(msgs)
      }
    </div>
    <a href="/contacts" class="btn btn-secondary">Cancel</a>
    <input type="submit" value="Save"/>
  </form>
}
```

### Display the form

The next step is to display the form to the user.

On `GET` requests, the form is displayed with an empty model for adding a new contact, or with an existing contact's data for editing.

```go
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
  model := NewModel()
  // If it's an edit request, populate the model with existing data.
  if id := r.PathValue("id"); id != "" {
    contact, ok, err := h.DB.Get(r.Context(), id)
    if err != nil {
      h.Log.Error("Failed to get contact", slog.String("id", id), slog.Any("error", err))
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
    if !ok {
      http.Redirect(w, r, "/contacts/edit", http.StatusSeeOther)
      return
    }
    model = ModelFromContact(contact)
  }
  h.DisplayForm(w, r, model)
}
```

### Handle form submission

When the form is submitted, the `POST` request is handled by parsing the form data and decoding it into the model using the Gorilla `schema` package.

If validation fails, the form is redisplayed with error messages.

```go
func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
  // Parse the form.
  err := r.ParseForm()
  if err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
  }

  var model Model

  // Decode the form.
  dec := schema.NewDecoder()
  dec.IgnoreUnknownKeys(true)
  err = dec.Decode(&model, r.PostForm)
  if err != nil {
    h.Log.Warn("Failed to decode form", slog.Any("error", err))
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
  }

  // Validate the input.
  if len(model.Validate()) > 0 {
    h.DisplayForm(w, r, model)
    return
  }

  // Save the contact.
  id := r.PathValue("id")
  if id == "" {
    id = ksuid.New().String()
  }
  contact := db.NewContact(id, model.Name, model.Email)
  if err = h.DB.Save(r.Context(), contact); err != nil {
    h.Log.Error("Failed to save contact", slog.String("id", id), slog.Any("error", err))
    model.Error = "Failed to save the contact. Please try again."
    h.DisplayForm(w, r, model)
    return
  }

  // Redirect back to the contact list.
  http.Redirect(w, r, "/contacts", http.StatusSeeOther)
}
```

## Example project

The `crud` project is a simple web application that allows users to manage contacts. It demonstrates how to handle forms, validation, and database interactions using Go's standard library and the Gorilla schema package.

For full example code, see `./examples/crud` in `github.com/a-h/templ`.

- `main.go`: The entrypoint of the application.
- `db`: Contains database logic, including models and database operations.
- `routes`: Contains the HTTP handlers for different routes.
- `layout`: Contains the common layout for all pages.
- `static`: Contains static assets like CSS, JavaScript, and images.

### Entrypoint

The `main.go` file is the entrypoint of the application.

A common pattern in Go applications is to define a `run` function that can return an error to the main function.

```go title="main.go"
var dbURI = "file:data.db?mode=rwc"
var addr = "localhost:8080"

func main() {
  log := slog.Default()
  ctx := context.Background()
  if err := run(ctx, log); err != nil {
    log.Error("Failed to run server", slog.Any("error", err))
    os.Exit(1)
  }
}
```

The `run` function first initializes the database connection.

```go title="main.go"
pool, err := sqlitex.NewPool(dbURI, sqlitex.PoolOptions{})
if err != nil {
    log.Error("Failed to open database", slog.Any("error", err))
    return err
}
store := sqlitekv.New(pool)
if err := store.Init(ctx); err != nil {
    log.Error("Failed to initialize store", slog.Any("error", err))
    return err
}
db := db.New(store)
```

Next, it sets up the HTTP server with routes for the home page, contacts listing, and contact management (add/edit/delete).

```go title="main.go"
mux := http.NewServeMux()

homeHandler := home.NewHandler()
mux.Handle("/", homeHandler)

ch := contacts.NewHandler(log, db)
mux.Handle("/contacts", ch)

ceh := contactsedit.NewHandler(log, db)
mux.Handle("/contacts/edit", ceh)
mux.Handle("/contacts/edit/{id}", ceh)

cdh := contactsdelete.NewHandler(log, db)
mux.Handle("/contacts/delete/{id}", cdh)
```

The `static` directory contains scripts, CSS and images, and is served using Go's built in file serving handler.


```go title="main.go"
mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
```

Finally, the server is started on the specified address and port.

```go title="main.go"
log.Info("Starting server", slog.String("address", addr))
return http.ListenAndServe(addr, mux)
```

### Listing contacts

The route at `/contacts` renders a list of contacts, allowing users to view existing contacts and navigate to forms for adding, editing or deleting contacts.

The handler collects the list of contacts from the database, and passes it to the `View`, wrapping it all in `layout.Handler` so that the page is rendered with the common layout.

It's common practice to create a constructor function for the handler, and to define methods on the handler struct for each HTTP method that the handler supports to separate behaviour.

```go title="routes/contacts/handler.go"
func NewHandler(log *slog.Logger, db *db.DB) http.Handler {
  return &Handler{
    Log: log,
    DB:  db,
  }
}

type Handler struct {
  Log *slog.Logger
  DB  *db.DB
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  switch r.Method {
    case http.MethodGet:
      h.Get(w, r)
    default:
      http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
  }
}
```

The `Get` method retrieves the list of contacts from the database and passes it to the `View` template for rendering, using a standard layout.

```go title="routes/contacts/handler.go"
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
  contacts, err := h.DB.List(r.Context())
  if err != nil {
    h.Log.Error("Failed to list contacts", slog.Any("error", err))
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  v := layout.Handler(View(contacts))
  v.ServeHTTP(w, r)
}
```

The view is a simple table containing a bit of logic to display "No contacts" if the list is empty, and links to edit or delete each contact.

It's common to break down a page into smaller components, so the `ContactsList` component is used to display the list of contacts, and is called from the `View` template.

```templ title="routes/contacts/view.templ"
templ View(contacts []db.Contact) {
  <h1>Contacts</h1>
  <ul>
    <li><a href="/contacts/edit" hx-boost="true">Add contact</a></li>
  </ul>
  if len(contacts) == 0 {
    <p>No contacts</p>
  } else {
    @ContactList(contacts)
  }
}

templ ContactList(contacts []db.Contact) {
  <table class="table">
  <tr>
    <th>
      Name
    </th>
    <th>
      Email
    </th>
    <th>
      Actions
    </th>
  </tr>
  for _, contact := range contacts {
    <tr>
      <td>{ contact.Name }</td>
      <td>{ contact.Email }</td>
      <td>
        <a href={ fmt.Sprintf("/contacts/edit/%s", url.PathEscape(contact.ID)) } hx-boost="true">Edit</a>
        <a href={ fmt.Sprintf("/contacts/delete/%s", url.PathEscape(contact.ID)) } hx-boost="true">Delete</a>
      </td>
    </tr>
  }
  </table>
}
```

:::tip
For simple views, there's no need to create a view model (a struct that defines the data that will be displayed) and you can pass the data directly, but for more complex views or when you need to pass additional data to the template, it's usually clearer to define a view model.
:::


### Layout

The `layout` package provides a common structure for all pages, including links to static assets like CSS and JavaScript files.

The `content` component passed into the `Page` template is replaced with the specific content for each page. Multiple function arguments or structs can be passed to the `Page` template to enable multiple slots for content.

```templ title="layout/page.templ"
package layout

templ Page(content templ.Component) {
  <!DOCTYPE html>
  <html>
    <head>
      <script src="/static/htmx.min.js"></script>
      <link rel="stylesheet" href="/static/bootstrap.css"/>
    </head>
    <body class="container">
      @content
    </body>
  </html>
}
```

A small helper function wraps the `Page` template to create an HTTP handler that can be used in routes.

```go title="layout/layout.go"
func Handler(content templ.Component) http.Handler {
  return templ.Handler(Page(content))
}
```

### Adding and editing contacts

The `/contacts/edit` route is used for both adding a new contact and editing an existing one. The handler checks if an ID is provided in the URL to determine whether to create a new contact or edit an existing one.

For `Get` requests, the handler retrieves the contact if an ID is provided, or initializes a new model for adding a contact. The `DisplayForm` method renders the form using the `View` template.

```go title="routes/contactsedit/handler.go"
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
  // Read the ID from the URL.
  id := r.PathValue("id")
  model := NewModel()
  if id != "" {
    // Get the existing contact from the database and populate the form.
    contact, ok, err := h.DB.Get(r.Context(), id)
    if err != nil {
      h.Log.Error("Failed to get contact", slog.String("id", id), slog.Any("error", err))
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return err
    }
    if !ok {
      http.Redirect(w, r, "/contacts/edit", http.StatusSeeOther)
      return
    }
    model = ModelFromContact(contact)
  }
  h.DisplayForm(w, r, model)
}
```

Note that the `ModelFromContact` function is used to convert a `db.Contact` into a view model (`Model`) that can be used to populate the form fields.


The `DisplayForm` method handles rendering the form view and is used by both the `Get` and `Post` methods. It uses the `layout.Handler` to ensure that the form is rendered within the common layout of the application.

```go title="routes/contactsedit/handler.go"
func (h *Handler) DisplayForm(w http.ResponseWriter, r *http.Request, m Model) {
  layout.Handler(View(m)).ServeHTTP(w, r)
}
```

For `Post` requests, the handler parses the form data into the model, validates it, and saves the contact to the database. If validation fails, it redisplays the form with error messages.


```go title="routes/contactsedit/handler.go"
func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
  // Parse the form.
  err := r.ParseForm()
  if err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
  }

  var model Model

  // Decode the form.
  dec := schema.NewDecoder()
  dec.IgnoreUnknownKeys(true)
  err = dec.Decode(&model, r.PostForm)
  if err != nil {
    h.Log.Warn("Failed to decode form", slog.Any("error", err))
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
  }

  // Validate the input.
  if len(model.Validate()) > 0 {
    h.DisplayForm(w, r, model)
    return
  }

  // Save the contact.
  id := r.PathValue("id")
  if id == "" {
    id = ksuid.New().String()
  }
  contact := db.NewContact(id, model.Name, model.Email)
  if err = h.DB.Save(r.Context(), contact); err != nil {
    h.Log.Error("Failed to save contact", slog.String("id", id), slog.Any("error", err))
    model.Error = "Failed to save the contact. Please try again."
    h.DisplayForm(w, r, model)
    return
  }

  // Redirect back to the contact list.
  http.Redirect(w, r, "/contacts", http.StatusSeeOther)
}
```

The validation is carried out by a `Validate` method on the model, which checks for required fields and returns a list of errors if any are found. This allows for complex validation logic to be encapsulated within the model itself.

```go title=./routes/contactsedit/model.go
func NewModel() Model {
  return Model{
    Initial: true,
  }
}

func ModelFromContact(contact db.Contact) (m Model) {
  return Model{
    Initial: true,
    Name:    contact.Name,
    Email:   contact.Email,
  }
}

type Model struct {
  Initial          bool
  SubmitButtonText string

  Name  string
  Email string
  Error string
}

func (m *Model) ValidateName() (msgs []string) {
  if m.Initial {
    return
  }
  if m.Name == "" {
    msgs = append(msgs, "Name is required")
  }
  return msgs
}

func (m *Model) NameHasError() bool {
  return len(m.ValidateName()) > 0
}

func (m *Model) ValidateEmail() (msgs []string) {
  if m.Initial {
    return
  }
  if m.Email == "" {
    return append(msgs, "Email is required")
  }
  if !strings.Contains(m.Email, "@") {
    msgs = append(msgs, "Email is invalid")
  }
  return msgs
}

func (m *Model) EmailHasError() bool {
  return len(m.ValidateEmail()) > 0
}

func (m *Model) Validate() (msgs []string) {
  if m.Initial {
    return
  }
  msgs = append(msgs, m.ValidateName()...)
  msgs = append(msgs, m.ValidateEmail()...)
  return msgs
}
```

The view for the contact form is defined in `view.templ`, which uses templ to render the form fields and any validation errors.

```templ title=./routes/contact/sedit/view.templ
package contactsedit

templ View(m Model) {
  <h1>Add Contact</h1>
  <ul>
    <li><a href="/contacts" hx-boost="true">Back to Contacts</a></li>
  </ul>
  <form id="form" method="post" hx-boost="true">
    <div id="name-group" class={ "form-group", templ.KV("has-error", m.NameHasError()) }>
      <label for="name">Name</label>
      <input type="text" id="name" name="name" class="form-control" placeholder="Name" value={ m.Name }/>
    </div>
    <div id="email-group" class={ "form-group", templ.KV("has-error", m.EmailHasError()) }>
      <label for="email">Email</label>
      <input type="email" id="email" name="email" class="form-control" placeholder="Email" value={ m.Email }/>
    </div>
    <div id="validation">
      if m.Error != "" {
        <p class="error">{ m.Error }</p>
      }
      if msgs := m.Validate(); len(msgs) > 0 {
        @ValidationMessages(msgs)
      }
    </div>
    <a href="/contacts" class="btn btn-secondary">Cancel</a>
    <input type="submit" value="Save"/>
  </form>
}

templ ValidationMessages(msgs []string) {
  if len(msgs) > 0 {
    <div class="invalid-feedback">
      <ul>
        for _, msg := range msgs {
          <li class="error">{ msg }</li>
        }
      </ul>
    </div>
  }
}
```

:::note
The `hx-boost="true"` attribute on the form enables htmx to handle the form submission via AJAX, allowing for a smoother user experience without full page reloads.
:::

### Deleting a contact

The `/contacts/delete/{id}` route handles the deletion of a contact. The handler retrieves the contact by ID and displays a confirmation form.

After the user confirms the deletion, the contact is removed from the database and the user is redirected back to the contact list.

```go title=./routes/contactsdelete/handler.go
type Handler struct {
  Log *slog.Logger
  DB  *db.DB
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  switch r.Method {
  case http.MethodGet:
    h.Get(w, r)
  case http.MethodPost:
    h.Post(w, r)
  default:
    http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
  }
}

func NewModel(name string) Model {
  return Model{
    Name: name,
  }
}

type Model struct {
  Name string
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
  // Read the ID from the URL.
  id := r.PathValue("id")
  if id == "" {
    http.Redirect(w, r, "/contacts", http.StatusSeeOther)
    return
  }
  // Get the existing contact from the database.
  contact, ok, err := h.DB.Get(r.Context(), id)
  if err != nil {
    h.Log.Error("Failed to get contact", slog.String("id", id), slog.Any("error", err))
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  if !ok {
    http.Redirect(w, r, "/contacts", http.StatusSeeOther)
    return
  }
  h.DisplayForm(w, r, NewModel(contact.Name))
}

func (h *Handler) DisplayForm(w http.ResponseWriter, r *http.Request, m Model) {
  layout.Handler(View(m)).ServeHTTP(w, r)
}

func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
  id := r.PathValue("id")
  if id == "" {
    http.Redirect(w, r, "/contacts", http.StatusSeeOther)
    return
  }

  // Delete the contact from the database.
  err := h.DB.Delete(r.Context(), id)
  if err != nil {
    h.Log.Error("Failed to delete contact", slog.String("id", id), slog.Any("error", err))
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }

  // Redirect back to the contact list.
  http.Redirect(w, r, "/contacts", http.StatusSeeOther)
}
```

The view for the delete confirmation is straightforward, displaying the contact's name and asking for confirmation before deletion.

```templ title=./routes/contactsdelete/view.templ
templ View(m Model) {
  <h1>Delete</h1>
  <p>
    Are you sure you want to delete <strong>{ m.Name }</strong>?
  </p>
  <form id="form" method="post" hx-boost="true">
    @csrf.CSRF()
    <a href="/contacts" class="btn btn-secondary">Cancel</a>
    <input type="submit" value="Delete"/>
  </form>
}
```
