---
sidebar_position: 1
---

# Introduction

## Overview of templ

`todo`

## Features and benefits

`todo`

```html
templ PersonTemplate(p Person) {
	<div>
	    for _, v := range p.Addresses {
		    {! AddressTemplate(v) }
	    }
	</div>
}
```

```go
templ PersonTemplate(p Person) {
	<div>
	    for _, v := range p.Addresses {
		    {! AddressTemplate(v) }
	    }
	</div>
}
```
