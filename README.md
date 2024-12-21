<meta name="author" content="Bougou">
<meta name="description" content="Go IPMI library">
<meta name="keywords" content="ipmi, go, golang, bmc">

# [go-ipmi](https://github.com/bougou/go-ipmi)

[`go-ipmi`](https://github.com/bougou/go-ipmi) is a pure golang native IPMI library. It DOES NOT wraps `ipmitool`.

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
	// Support local mode client if runs directly on linux
	// client, err := ipmi.NewOpenClient()
	if err != nil {
		panic(err)
	}

	// you can optionally open debug switch
	// client.WithDebug(true)

	// you can set interface type, enum range: open/lan/lanplus/tool, default open
	// client.WithInterface(ipmi.InterfaceLanplus)

	// !!! Note !!!,
	// From v0.6.0, all IPMI command methods of the Client accept a context as the first argument.
	ctx := context.Background()

	// Connect will create an authenticated session for you.
	if err := client.Connect(ctx); err != nil {
		panic(err)
	}

	// Now you can execute other IPMI commands that need authentication.

	res, err := client.GetDeviceID(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println(res.Format())

	selEntries, err := client.GetSELEntries(ctx, 0)
	if err != nil {
		panic(err)
	}
	fmt.Println(ipmi.FormatSELs(selEntries, nil))
}
```

## `goipmi` binary

The goipmi is a binary tool which provides the same command usages like ipmitool. The goipmi calls go-impi library underlying.

The purpose of creating goipmi tool was not intended to substitute ipmitool.
It was not strictly crafted, and was just used to verify the correctness of go-ipmi library.

## Functions Comparison with ipmitool

Each command defined in the IPMI specification is a pair of request/response messages.
These IPMI commands are implemented as methods of the `ipmi.Client` struct in this library.

Some `ipmitool` cmdline usages are implemented by calling just one IPMI command,
but others are not. Like `ipmitool sdr list`, it's a loop of `GetSDR` IPMI command.

So this library also implements some methods that are not IPMI commands defined
in IPMI specification, but just some common helpers, like `GetSDRs` to get all SDRs.
These methods are marked with an asterisk `(*)` after the method name in the following docs.

The implementation logic of IPMI commands is almost same. See [Contributing](./CONTRIBUTING.md)

> More commands are ongoing ...

### IPM Device Global Commands


| Method                             | Status             | corresponding ipmitool usage  |
| ---------------------------------- | ------------------ | ----------------------------- |
| GetDeviceID                        | :white_check_mark: | mc info                       |
| ColdReset                          | :white_check_mark: | mc reset cold                 |
| WarmReset                          | :white_check_mark: | mc reset warm                 |
| GetSelfTestResults                 | :white_check_mark: | mc selftest, chassis selftest |
| ManufacturingTestOn                | :white_check_mark: |
| SetACPIPowerState                  | :white_check_mark: |
| GetACPIPowerState                  | :white_check_mark: |
| GetDeviceGUID                      | :white_check_mark: |
| GetNetFnSupport                    | :white_check_mark: |
| GetCommandSupport                  | :white_check_mark: |
| GetCommandSubfunctionSupport       |                    |
| GetConfigurableCommands            | :white_check_mark: |
| GetConfigurableCommandSubfunctions |                    |
| SetCommandEnables                  |                    |
| GetCommandEnables                  | :white_check_mark: |
| GetCommandSubfunctionsEnables      | :white_check_mark: |
| GetSubfunctionsEnables             |                    |
| GetOEMNetFnIanaSupport             |                    |

### BMC Watchdog Timer Commands

| Method             | Status             | corresponding ipmitool usage |
| ------------------ | ------------------ | ---------------------------- |
| ResetWatchdogTimer | :white_check_mark: | mc watchdog reset            |
| SetWatchdogTimer   | :white_check_mark: |
| GetWatchdogTimer   | :white_check_mark: | mc watchdog get              |

### BMC Device and Messaging Commands

| Method                         | Status             | corresponding ipmitool usage |
| ------------------------------ | ------------------ | ---------------------------- |
| SetBMCGlobalEnables            | :white_check_mark: |
| GetBMCGlobalEnables            | :white_check_mark: |
| ClearMessageFlags              | :white_check_mark: |
| GetMessageFlags                | :white_check_mark: |
| EnableMessageChannelReceive    | :white_check_mark: |
| GetMessage                     | :white_check_mark: |
| SendMessage                    | :white_check_mark: |
| ReadEventMessageBuffer         | :white_check_mark: |
| GetBTInterfaceCapabilities     |                    |
| GetSystemGUID                  | :white_check_mark: | mc guid                      |
| SetSystemInfoParams            |                    |
| GetSystemInfoParams            |                    |
| GetChannelAuthCapabilities     | :white_check_mark: |
| GetSessionChallenge            | :white_check_mark: |
| ActivateSession                | :white_check_mark: |
| SetSessionPrivilegeLevel       | :white_check_mark: |
| CloseSession                   | :white_check_mark: |
| GetSessionInfo                 | :white_check_mark: | session info                 |
| GetAuthCode                    | :white_check_mark: |
| SetChannelAccess               | :white_check_mark: | channel setaccess            |
| GetChannelAccess               | :white_check_mark: | channel info/getaccess       |
| GetChannelInfo                 | :white_check_mark: | channel info                 |
| SetUserAccess                  | :white_check_mark: |
| GetUserAccess                  | :white_check_mark: | user summary                 |
| GetUsers (*)                   | :white_check_mark: | user list                    |
| SetUsername                    | :white_check_mark: | user set name                |
| DisableUser (*)                | :white_check_mark: | user disable                 |
| EnableUser (*)                 | :white_check_mark: | user enable                  |
| GetUsername                    | :white_check_mark: |
| SetUserPassword                | :white_check_mark: | user set password            |
| TestUserPassword (*)           | :white_check_mark: | user test                    |
| ActivatePayload                |                    |
| DeactivatePayload              |                    |
| GetPayloadActivationStatus     |                    |
| GetPayloadInstanceInfo         |                    |
| SetUserPayloadAccess           |                    |
| GetUserPayloadAccess           |                    | sol payload status           |
| GetChannelPayloadSupport       |                    |
| GetChannelPayloadVersion       |                    |
| GetChannelOEMPayloadInfo       |                    |
| MasterWriteRead                |                    |
| GetChannelCipherSuites         | :white_check_mark: |
| SuspendOrResumeEncryption      |                    |
| SetChannelCipherSuites         |                    |
| GetSystemInterfaceCapabilities | :white_check_mark: |

### Chassis Device Commands

| Method                    | Status             | corresponding ipmitool usage                      |
| ------------------------- | ------------------ | ------------------------------------------------- |
| GetChassisCapabilities    | :white_check_mark: |
| GetChassisStatus          | :white_check_mark: | chassis status, chassis power status              |
| ChassisControl            | :white_check_mark: | chassis power on/off/cycle/reset/diag/soft        |
| ChassisReset              | :white_check_mark: |
| ChassisIdentify           | :white_check_mark: | chassis identify                                  |
| SetChassisCapabilities    | :white_check_mark: |
| SetPowerRestorePolicy     | :white_check_mark: | chassis policy list/always-on/previous/always-off |
| GetSystemRestartCause     | :white_check_mark: | chassis restart_cause                             |
| SetSystemBootOptions      | :white_check_mark: | chassis bootparam set                             |
| SetBootParamBootFlags (*) | :white_check_mark: | chassis bootdev                                   |
| GetSystemBootOptions      | :white_check_mark: | chassis bootparam get                             |
| SetFrontPanelEnables      | :white_check_mark: |
| SetPowerCycleInterval     | :white_check_mark: |
| GetPOHCounter             | :white_check_mark: | chassis poh                                       |

### Event Commands

| Method               | Status             | corresponding ipmitool usage |
| -------------------- | ------------------ | ---------------------------- |
| SetEventReceiver     | :white_check_mark: |
| GetEventReceiver     | :white_check_mark: |
| PlatformEventMessage | :white_check_mark: |

### PEF and Alerting Commands

| Method                  | Status             | corresponding ipmitool usage |
| ----------------------- | ------------------ | ---------------------------- |
| GetPEFCapabilities      | :white_check_mark: | pef capabilities             |
| ArmPEFPostponeTimer     |                    |
| SetPEFConfigParams      |                    |
| GetPEFConfigParams      |                    |
| SetLastProcessedEventId |                    |
| GetLastProcessedEventId |                    |
| AlertImmediate          |                    |
| PEFAck                  |                    |

### Sensor Device Commands

| Method                         | Status             | corresponding ipmitool usage |
| ------------------------------ | ------------------ | ---------------------------- |
| GetDeviceSDRInfo               | :white_check_mark: |
| GetDeviceSDR                   | :white_check_mark: |
| ReserveDeviceSDRRepo           | :white_check_mark: |
| GetSensorReadingFactors        | :white_check_mark: |
| SetSensorHysteresis            | :white_check_mark: |
| GetSensorHysteresis            | :white_check_mark: |
| SetSensorThresholds            | :white_check_mark: |
| GetSensorThresholds            | :white_check_mark: |
| SetSensorEventEnable           |                    |
| GetSensorEventEnable           | :white_check_mark: |
| RearmSensorEvents              |                    |
| GetSensorEventStatus           | :white_check_mark: |
| GetSensorReading               | :white_check_mark: |
| SetSensorType                  | :white_check_mark: |
| GetSensorType                  | :white_check_mark: |
| SetSensorReadingAndEventStatus | :white_check_mark: |
| GetSensors (*)                 | :white_check_mark: | sensor list, sdr type        |
| GetSensorByID (*)              | :white_check_mark: |                              |
| GetSensorByName (*)            | :white_check_mark: | sensor get                   |

### FRU Device Commands

| Method                  | Status             | corresponding ipmitool usage |
| ----------------------- | ------------------ | ---------------------------- |
| GetFRUInventoryAreaInfo | :white_check_mark: |
| ReadFRUData             | :white_check_mark: |
| WriteFRUData            | :white_check_mark: |
| GetFRU (*)              | :white_check_mark: | fru print                    |
| GetFRUs (*)             | :white_check_mark: | fru print                    |


### SDR Device Commands

| Method                 | Status             | corresponding ipmitool usage |
| ---------------------- | ------------------ | ---------------------------- |
| GetSDRRepoInfo         | :white_check_mark: | sdr info                     |
| GetSDRRepoAllocInfo    | :white_check_mark: | sdr info                     |
| ReserveSDRRepo         |                    |
| GetSDR                 | :white_check_mark: |                              |
| GetSDRs (*)            | :white_check_mark: |                              |
| GetSDRBySensorID (*)   | :white_check_mark: |                              |
| GetSDRBySensorName (*) | :white_check_mark: |
| AddSDR                 |                    |
| PartialAddSDR          |                    |
| DeleteSDR              |                    |
| ClearSDRRepo           |                    |
| GetSDRRepoTime         |                    |
| SetSDRRepoTime         |                    |
| EnterSDRRepoUpdateMode |                    |
| ExitSDRRepoUpdateMode  |                    |
| RunInitializationAgent |                    |

### SEL Device Commands

| Method              | Status             | corresponding ipmitool usage |
| ------------------- | ------------------ | ---------------------------- |
| GetSELInfo          | :white_check_mark: | sel info                     |
| GetSELAllocInfo     | :white_check_mark: | sel info                     |
| ReserveSEL          | :white_check_mark: |
| GetSELEntry         | :white_check_mark: |
| AddSELEntry         | :white_check_mark: |
| PartialAddSELEntry  |                    |
| DeleteSELEntry      | :white_check_mark: |
| ClearSEL            | :white_check_mark: | sel clear                    |
| GetSELTime          | :white_check_mark: |
| SetSELTime          | :white_check_mark: |
| GetAuxLogStatus     |                    |
| SetAuxLogStatus     |                    |
| GetSELTimeUTCOffset | :white_check_mark: |
| SetSELTimeUTCOffset | :white_check_mark: |

### LAN Device Commands

| Method             | Status             | corresponding ipmitool usage |
| ------------------ | ------------------ | ---------------------------- |
| SetLanConfigParams |                    |
| GetLanConfigParams | :white_check_mark: |
| GetLanConfig (*)   | :white_check_mark: | lan print                    |
| SuspendARPs        | :white_check_mark: |
| GetIpStatistics    | :white_check_mark: |

### Serial/Modem Device Commands

| Method                 | Status             | corresponding ipmitool usage |
| ---------------------- | ------------------ | ---------------------------- |
| SetSerialConfig        |                    |
| GetSerialConfig        |                    |
| SetSerialMux           |                    |
| GetTapResponseCodes    |                    |
| SetPPPTransmitData     |                    |
| GetPPPTransmitData     |                    |
| SendPPPPacket          |                    |
| GetPPPReceiveData      |                    |
| SerialConnectionActive |                    |
| Callback               |                    |
| SetUserCallbackOptions |                    |
| GetUserCallbackOptions |                    |
| SetSerialRoutingMux    |                    |
| SOLActivating          | :white_check_mark: |
| GetSOLConfigParams     | :white_check_mark: |
| SetSOLConfigParams     | :white_check_mark: |
| SOLInfo                | :white_check_mark: | sol info                     |

### Command Forwarding Commands

| Method          | Status | corresponding ipmitool usage |
| --------------- | ------ | ---------------------------- |
| Forwarded       |        |
| SetForwarded    |        |
| GetForwarded    |        |
| EnableForwarded |        |

### Bridge Management Commands (ICMB)

| Method                | Status | corresponding ipmitool usage |
| --------------------- | ------ | ---------------------------- |
| GetBridgeState        |        |
| SetBridgeState        |        |
| GetICMBAddress        |        |
| SetICMBAddress        |        |
| SetBridgeProxyAddress |        |
| GetBridgeStatistics   |        |
| GetICMBCapabilities   |        |
| ClearBridgeStatistics |        |
| GetBridgeProxyAddress |        |
| GetICMBConnectorInfo  |        |
| GetICMBConnectionID   |        |
| SendICMBConnectionID  |        |

### Discovery Commands (ICMB)

| Method              | Status | corresponding ipmitool usage |
| ------------------- | ------ | ---------------------------- |
| PrepareForDiscovery |        |
| GetAddresses        |        |
| SetDiscovered       |        |
| GetChassisDeviceId  |        |
| SetChassisDeviceId  |        |

### Bridging Commands (ICMB)

| Method        | Status | corresponding ipmitool usage |
| ------------- | ------ | ---------------------------- |
| BridgeRequest |        |
| BridgeMessage |        |

### Event Commands (ICMB)

| Method                 | Status | corresponding ipmitool usage |
| ---------------------- | ------ | ---------------------------- |
| GetEventCount          |        |
| SetEventDestination    |        |
| SetEventReceptionState |        |
| SendICMBEventMessage   |        |
| GetEventDestination    |        |
| GetEventReceptionState |        |


### Other Bridge Commands

| Method      | Status | corresponding ipmitool usage |
| ----------- | ------ | ---------------------------- |
| ErrorReport |        |

### DCMI Commands

| Method              | Status             | corresponding ipmitool usage |
| ------------------- | ------------------ | ---------------------------- |
| GetDCMIPowerReading | :white_check_mark: | dcmi power reading           |
| GetDCMIAssetTag     | :white_check_mark: | dcmi asset_tag               |

## Reference

- [Intelligent Platform Management Interface Specification Second Generation v2.0](https://www.intel.com/content/dam/www/public/us/en/documents/specification-updates/ipmi-intelligent-platform-mgt-interface-spec-2nd-gen-v2-0-spec-update.pdf)
- [Platform Management FRU Information Storage Definition](https://www.intel.com/content/dam/www/public/us/en/documents/specification-updates/ipmi-platform-mgt-fru-info-storage-def-v1-0-rev-1-3-spec-update.pdf)
- [PC SDRAM Serial Presence Detect (SPD) Specification](https://cdn.hackaday.io/files/10119432931296/Spdsd12b.pdf)
- [DCMI Group Extension Specification v1.5](https://www.intel.com/content/dam/www/public/us/en/documents/technical-specifications/dcmi-v1-5-rev-spec.pdf)
