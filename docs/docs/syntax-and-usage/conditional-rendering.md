---
sidebar_position: 3
---

# Conditional rendering

## if/else

templ uses standard Go `if`/`else` statements.

```templ
templ login(isLoggedIn bool) {
  if isLoggedIn {
    <div>Welcome back!</div>
  } else {
    <input name="login" type="button" value="Log in"/>
  }
}
```

## switch

templ uses standard Go `switch` statements.

```templ
templ userType(isLoggedIn bool) {
  switch p.Type {
    case "test":
      <span>{ "Test user" }</span>
    case "admin"
      <span>{ "Admin user" }</span>
    default:
      <span>{ "Unknown user" }</span>
  }
}
```

## Conditional attributes

Use an `if` statement within a templ element to optionally add attributes to elements.

```templ
templ conditionalAttribute() {
  <hr style="padding: 10px" 
    if true {
      class="itIsTrue"
    }
  />
}
```

```html
<hr style="padding: 10px" class="itIsTrue" />
```
