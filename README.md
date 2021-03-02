# paypayopa-sdk-go

paypayopa-sdk-go is a Go client library for accessing the [PayPay API](https://www.paypay.ne.jp/opa/doc/jp/v1.0/continuous_payments).

## Usage

```go
import "github.com/gotokatsuya/paypayopa-sdk-go/paypay"

func main() {
    pay, err := paypay.New("..")
    ...
}
```

## License

This library is distributed under the MIT license.
