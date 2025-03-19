# XJTU Payment Platform

Go module for XJTU Payment

### Usage
```go
import (
    "github.com/endaytrer/xjtupay"
)
info := xjtupay.PaymentInfo {
    ...
}

payment, err := xjtupay.InitPayment(true, info)

if err != nil {
    // deal with error
}

var redir_client *http.Client

switch t := payment.(type) {

case *xjtupay.PayWithPasswd:
    redir_client, err = t.Pay("888888")

case *xjtupay.PayWithoutPasswd:
    redir_client, err = t.Pay()
}
if err != nil {
    // deal with error
}

// deal with redirection client

```