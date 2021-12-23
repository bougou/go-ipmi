# go-ipmi

go-ipmi is a pure golang native IPMI library. It DOES NOT wraps `ipmitool`.

## Usage

```go
import github.com/bougou/go-ipmi

host := "10.0.0.1"
port := 623
username := "root"
password := "123456"

client, err := ipmi.NewClient(host, port, username, password)
if err != nil {
  panic(err)
}

// Connect will create an authenticated session for you.
if err := client.Connect(); err != nil {
  panic(err)
}

// Now you can executed other commands that need authentication.
res, err := client.SELList()
if err != nil {
  panic(err)
}
fmt.Println(res)
```

## commands / ipmitool

> More commands are ongoing ...
>
| Client Method       | ipmitool command   |
| ------------------- | ------------------ |
| GetSELInfo          | ipmitool sel info  |
| GetSELAllocInfo     | -                  |
| ClearSEL            | ipmitool sel clear |
| GetSDRRepoInfo      | ipmitool sdr info  |
| GetSDRRepoAllocInfo | -                  |
