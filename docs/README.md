<meta name="author" content="Bougou">
<meta name="description" content="Go IPMI library">
<meta name="keywords" content="ipmi, go, golang, bmc">
<meta name="google-site-verification" content="Ejz48wAig8QjFaggJoluq4crKN7x7Jbi_VnEqFXQIhs" />

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

	// Connect will create an authenticated session for you.
	if err := client.Connect(); err != nil {
		panic(err)
	}

	// Now you can execute other IPMI commands that need authentication.

	res, err := client.GetDeviceID()
	if err != nil {
		panic(err)
	}
	fmt.Println(res.Format())

	selEntries, err := client.GetSELEntries(0)
	if err != nil {
		panic(err)
	}
	fmt.Println(ipmi.FormatSELs(selEntries, nil))
}
```

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


| Method                             | Status  | corresponding ipmitool usage  |
| ---------------------------------- | ------- | ----------------------------- |
| GetDeviceID                        | &check; | mc info                       |
| ColdReset                          | &check; | mc reset cold                 |
| WarmReset                          | &check; | mc reset warm                 |
| GetSelfTestResults                 | &check; | mc selftest, chassis selftest |
| ManufacturingTestOn                | &check; |
| SetACPIPowerState                  | &check; |
| GetACPIPowerState                  | &check; |
| GetDeviceGUID                      | &check; |
| GetNetFnSupport                    | &check; |
| GetCommandSupport                  | &check; |
| GetCommandSubfunctionSupport       |         |
| GetConfigurableCommands            | &check; |
| GetConfigurableCommandSubfunctions |         |
| SetCommandEnables                  |         |
| GetCommandEnables                  | &check; |
| GetCommandSubfunctionsEnables      | &check; |
| GetSubfunctionsEnables             |         |
| GetOEMNetFnIanaSupport             |         |

### BMC Watchdog Timer Commands

| Method             | Status  | corresponding ipmitool usage |
| ------------------ | ------- | ---------------------------- |
| ResetWatchdogTimer | &check; | mc watchdog reset            |
| SetWatchdogTimer   | &check; |
| GetWatchdogTimer   | &check; | mc watchdog get              |

### BMC Device and Messaging Commands

| Method                         | Status  | corresponding ipmitool usage |
| ------------------------------ | ------- | ---------------------------- |
| SetBMCGlobalEnables            | &check; |
| GetBMCGlobalEnables            | &check; |
| ClearMessageFlags              | &check; |
| GetMessageFlags                | &check; |
| EnableMessageChannelReceive    | &check; |
| GetMessage                     | &check; |
| SendMessage                    | &check; |
| ReadEventMessageBuffer         | &check; |
| GetBTInterfaceCapabilities     |         |
| GetSystemGUID                  | &check; | mc guid                      |
| SetSystemInfoParameters        |         |
| GetSystemInfoParameters        |         |
| GetChannelAuthCapabilities     | &check; |
| GetSessionChallenge            | &check; |
| ActivateSession                | &check; |
| SetSessionPrivilegeLevel       | &check; |
| CloseSession                   | &check; |
| GetSessionInfo                 | &check; | session info                 |
| GetAuthCode                    | &check; |
| SetChannelAccess               | &check; | channel setaccess            |
| GetChannelAccess               | &check; | channel info/getaccess       |
| GetChannelInfo                 | &check; | channel info                 |
| SetUserAccess                  | &check; |
| GetUserAccess                  | &check; | user summary                 |
| GetUsers (*)                   | &check; | user list                    |
| SetUsername                    | &check; | user set name                |
| DisableUser (*)                | &check; | user disable                 |
| EnableUser (*)                 | &check; | user enable                  |
| GetUsername                    | &check; |
| SetUserPassword                | &check; | user set password            |
| TestUserPassword (*)           | &check; | user test                    |
| ActivatePayload                |         |
| DeactivatePayload              |         |
| GetPayloadActivationStatus     |         |
| GetPayloadInstanceInfo         |         |
| SetUserPayloadAccess           |         |
| GetUserPayloadAccess           |         |
| GetChannelPayloadSupport       |         |
| GetChannelPayloadVersion       |         |
| GetChannelOEMPayloadInfo       |         |
| MasterWriteRead                |         |
| GetChannelCipherSuites         | &check; |
| SuspendOrResumeEncryption      |         |
| SetChannelCipherSuites         |         |
| GetSystemInterfaceCapabilities | &check; |

### Chassis Device Commands

| Method                    | Status  | corresponding ipmitool usage                      |
| ------------------------- | ------- | ------------------------------------------------- |
| GetChassisCapabilities    | &check; |
| GetChassisStatus          | &check; | chassis status, chassis power status              |
| ChassisControl            | &check; | chassis power on/off/cycle/reset/diag/soft        |
| ChassisReset              | &check; |
| ChassisIdentify           | &check; | chassis identify                                  |
| SetChassisCapabilities    | &check; |
| SetPowerRestorePolicy     | &check; | chassis policy list/always-on/previous/always-off |
| GetSystemRestartCause     | &check; | chassis restart_cause                             |
| SetSystemBootOptions      | &check; | chassis bootparam set                             |
| SetBootParamBootFlags (*) | &check; | chassis bootdev                                   |
| GetSystemBootOptions      | &check; | chassis bootparam get                             |
| SetFrontPanelEnables      | &check; |
| SetPowerCycleInterval     | &check; |
| GetPOHCounter             | &check; | chassis poh                                       |

### Event Commands

| Method               | Status  | corresponding ipmitool usage |
| -------------------- | ------- | ---------------------------- |
| SetEventReceiver     | &check; |
| GetEventReceiver     | &check; |
| PlatformEventMessage | &check; |

### PEF and Alerting Commands

| Method                  | Status  | corresponding ipmitool usage |
| ----------------------- | ------- | ---------------------------- |
| GetPEFCapabilities      | &check; | pef capabilities             |
| ArmPEFPostponeTimer     |         |
| SetPEFConfigParameters  |         |
| GetPEFConfigParameters  |         |
| SetLastProcessedEventId |         |
| GetLastProcessedEventId |         |
| AlertImmediate          |         |
| PEFAck                  |         |

### Sensor Device Commands

| Method                         | Status  | corresponding ipmitool usage |
| ------------------------------ | ------- | ---------------------------- |
| GetDeviceSDRInfo               | &check; |
| GetDeviceSDR                   | &check; |
| ReserveDeviceSDRRepo           | &check; |
| GetSensorReadingFactors        | &check; |
| SetSensorHysteresis            | &check; |
| GetSensorHysteresis            | &check; |
| SetSensorThresholds            | &check; |
| GetSensorThresholds            | &check; |
| SetSensorEventEnable           |         |
| GetSensorEventEnable           | &check; |
| RearmSensorEvents              |         |
| GetSensorEventStatus           | &check; |
| GetSensorReading               | &check; |
| SetSensorType                  | &check; |
| GetSensorType                  | &check; |
| SetSensorReadingAndEventStatus | &check; |
| GetSensors (*)                 | &check; | sensor list                  |
| GetSensorByID (*)              | &check; |                              |
| GetSensorByName (*)            | &check; | sensor get                   |

### FRU Device Commands

| Method                  | Status  | corresponding ipmitool usage |
| ----------------------- | ------- | ---------------------------- |
| GetFRUInventoryAreaInfo | &check; |
| ReadFRUData             | &check; |
| WriteFRUData            | &check; |
| GetFRU (*)              | &check; | fru print                    |
| GetFRUs (*)             | &check; | fru print                    |


### SDR Device Commands

| Method                 | Status  | corresponding ipmitool usage |
| ---------------------- | ------- | ---------------------------- |
| GetSDRRepoInfo         | &check; | sdr info                     |
| GetSDRRepoAllocInfo    | &check; | sdr info                     |
| ReserveSDRRepo         |         |
| GetSDR                 | &check; |                              |
| GetSDRs (*)            | &check; |                              |
| GetSDRBySensorID (*)   | &check; |                              |
| GetSDRBySensorName (*) | &check; |
| AddSDR                 |         |
| PartialAddSDR          |         |
| DeleteSDR              |         |
| ClearSDRRepo           |         |
| GetSDRRepoTime         |         |
| SetSDRRepoTime         |         |
| EnterSDRRepoUpdateMode  |         |
| ExitSDRRepoUpdateMode  |         |
| RunInitializationAgent |         |

### SEL Device Commands

| Method              | Status  | corresponding ipmitool usage |
| ------------------- | ------- | ---------------------------- |
| GetSELInfo          | &check; | sel info                     |
| GetSELAllocInfo     | &check; | sel info                     |
| ReserveSEL          | &check; |
| GetSELEntry         | &check; |
| AddSELEntry         | &check; |
| PartialAddSELEntry  |         |
| DeleteSELEntry      | &check; |
| ClearSEL            | &check; | sel clear                    |
| GetSELTime          | &check; |
| SetSELTime          | &check; |
| GetAuxLogStatus     |         |
| SetAuxLogStatus     |         |
| GetSELTimeUTCOffset | &check; |
| SetSELTimeUTCOffset | &check; |

### LAN Device Commands

| Method             | Status  | corresponding ipmitool usage |
| ------------------ | ------- | ---------------------------- |
| SetLanConfigParams |         |
| GetLanConfigParams | &check; |
| SuspendARPs        | &check; |
| GetIpStatistics    | &check; |

### Serial/Modem Device Commands

| Method                 | Status  | corresponding ipmitool usage |
| ---------------------- | ------- | ---------------------------- |
| SetSerialConfig        |         |
| GetSerialConfig        |         |
| SetSerialMux           |         |
| GetTapResponseCodes    |         |
| SetPPPTransmitData     |         |
| GetPPPTransmitData     |         |
| SendPPPPacket          |         |
| GetPPPReceiveData      |         |
| SerialConnectionActive |         |
| Callback               |         |
| SetUserCallbackOptions |         |
| GetUserCallbackOptions |         |
| SetSerialRoutingMux    |         |
| SOLActivating          | &check; |
| GetSOLConfigParams     | &check; |
| SetSOLConfigParams     | &check; |
| SOLInfo                | &check; | sol info                     |

### Command Forwarding Commands

| Method          | Status | corresponding ipmitool usage |
| --------------- | ------ | ---------------------------- |
| Forwarded        |        |
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

## Reference

- [Intelligent Platform Management Interface Specification Second Generation v2.0](https://www.intel.com/content/dam/www/public/us/en/documents/specification-updates/ipmi-intelligent-platform-mgt-interface-spec-2nd-gen-v2-0-spec-update.pdf)
- [Platform Management FRU Information Storage Definition](https://www.intel.com/content/dam/www/public/us/en/documents/specification-updates/ipmi-platform-mgt-fru-info-storage-def-v1-0-rev-1-3-spec-update.pdf)
