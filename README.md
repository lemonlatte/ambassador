# ambassador

This is a go package for binding all chat platforms

## Prerequisite

* Go 1.7+

## Example

```

import (
    "fmt"
    "github.com/lemonlatte/ambassador"
)

a := ambassador.New("facebook", "test-token")
messages, err := a.translate(req.Body)

if err != nil {
    fmt.Println(err)
}

for _, msg := range messages {
//  do something...
}

err := a.SendText("hello, world")

if err != nil {
    fmt.Println(err)
}

```
