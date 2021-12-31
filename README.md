# go-ipmi

go-ipmi is a pure golang native IPMI library. It DOES NOT wraps `ipmitool`.

## Usage

```go
import (
	"fmt"
	"github.com/bougou/go-ipmi"
)

func main() {
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

	// Now you can execute other commands that need authentication.
	selEntries, err := client.GetSELEntries(0)
	if err != nil {
		panic(err)
	}
	for _, sel := range selEntries {
		fmt.Println(sel)
	}
}
```

## Functions Comparision with ipmitool

> More is ongoing ...
>
| Client Method         | ipmitool cmdline                                      |
| --------------------- | ----------------------------------------------------- |
| GetSELInfo            | ipmitool sel info                                     |
| GetSELAllocInfo       | ipmitool sel info                                     |
| ClearSEL              | ipmitool sel clear                                    |
| GetSDRRepoInfo        | ipmitool sdr info                                     |
| GetSDRRepoAllocInfo   | ipmitool sdr info                                     |
| GetSDR                | ipmitool sdr get                                      |
| GetSDRs               | ipmitool sdr list/elist                               |
| GetChassisStatus      | ipmitool chassis status                               |
| GetChassisStatus      | ipmitool chassis power status                         |
| ChassisControl        | ipmitool chassis power on/off/cycle/reset/diag/soft   |
| ChassisIdentify       | ipmitool chassis identify                             |
| SetFrontPanelEnables  |
| SetPowerRestorePolicy | ipmitool chassis policy always-on/previous/always-off |
| GetSystemRestartCause | ipmitool chassis restart_cause                        |

## Reference

- [Intelligent Platform Management Interface Specification Second Generation v2.0](https://www.intel.com/content/dam/www/public/us/en/documents/specification-updates/ipmi-intelligent-platform-mgt-interface-spec-2nd-gen-v2-0-spec-update.pdf)
- [Platform Management FRU Information Storage Definition](https://www.intel.com/content/dam/www/public/us/en/documents/specification-updates/ipmi-platform-mgt-fru-info-storage-def-v1-0-rev-1-3-spec-update.pdf)
