# Manufacturer names

A BMC reports its manufacturer as an IANA Private Enterprise Number.
`GetDeviceID` returns it in `res.ManufacturerID`; `res.Format()` includes it
in the device summary. The `iana` package turns that number into the
registered organization name, using a snapshot of the
[IANA registry](https://www.iana.org/assignments/enterprise-numbers/)
embedded in the binary, so lookups work offline.

```go
import "github.com/bougou/go-ipmi/pkg/oem/iana"

if name, ok := iana.Lookup(res.ManufacturerID); ok {
	fmt.Printf("Manufacturer Name         : %s\n", name)
}
```

`Lookup` returns `ok == false` for numbers the snapshot does not know; what
to print then is up to you. Don't compare manufacturers using the string
`Lookup` returns: the registration for an ID can change between snapshots.
Compare the ID and use the name for display.

## Size and memory

Importing `iana` embeds a gzip copy of the registry, about 840 KB for
roughly 67,000 entries; a stripped binary grows by just under 1 MiB. The
first `Lookup` unpacks it into a map of about 5 MiB that lives for the rest
of the process and is shared across goroutines.

## Updating the snapshot

A program sees new registrations only when it is rebuilt with a newer
go-ipmi. In a checkout, `make update-iana-pen` fetches the current registry
and regenerates the embedded data. A monthly GitHub workflow does the same
and opens a pull request.
