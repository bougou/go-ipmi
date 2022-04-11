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

## Functions Comparision with ipmitool

Each command defined in the IPMI specification is a pair of request/response messages.
These IPMI commands are implemented as methods of the `ipmi.Client` struct in this library.

Some `ipmitool` cmdline usages are implemented by calling just one IPMI command,
but others are not. Like `ipmitool sdr list`, it's a loop of `GetSDR` IPMI command.

So this library also implements some methods that are not IPMI commands defined
in IPMI sepcification, but just some common helpers, like `GetSDRs` to get all SDRs.
These methods are marked with an asterisk (*)` after the method name in the following docs.

The implementation logic of IPMI commands is almost same. See [Contributing](./CONTRIBUTING.md)

> More commmands are ongoing ...

### IPM Device Global Commands


| Method                             | Status | corresponding ipmitool usage  |
| ---------------------------------- | ------ | ----------------------------- |
| GetDeviceID                        | √      | mc info                       |
| ColdReset                          | √      | mc reset cold                 |
| WarmReset                          | √      | mc reset warm                 |
| GetSelfTestResults                 | √      | mc selftest, chassis selftest |
| ManufacturingTestOn                | √      |
| SetACPIPowerState                  | √      |
| GetACPIPowerState                  | √      |
| GetDeviceGUID                      | √      |
| GetNetFnSupport                    | √      |
| GetCommandSupport                  | √      |
| GetCommandSubfunctionSupport       |        |
| GetConfigurableCommands            | √      |
| GetConfigurableCommandSubfunctions |        |
| SetCommandEnables                  |        |
| GetCommandEnables                  | √      |
| GetCommandSubfunctionsEnables      | √      |
| GetSubfunctionsEnables             |        |
| GetOEMNetFnIanaSupport             |        |

### BMC Watchdog Timer Commands

| Method             | Status | corresponding ipmitool usage |
| ------------------ | ------ | ---------------------------- |
| ResetWatchdogTimer | √      | mc watchdog reset            |
| SetWatchdogTimer   | √      |
| GetWatchdogTimer   | √      | mc watchdog get              |

### BMC Device and Messaging Commands

| Method                         | Status | corresponding ipmitool usage |
| ------------------------------ | ------ | ---------------------------- |
| SetBMCGlobalEnables            | √      |
| GetBMCGlobalEnables            | √      |
| ClearMessageFlags              | √      |
| GetMessageFlags                | √      |
| EnableMessageChannelReceive    | √      |
| GetMessage                     | √      |
| SendMessage                    | √      |
| ReadEventMessageBuffer         | √      |
| GetBTInterfaceCapabilities     |        |
| GetSystemGUID                  | √      | mc guid                      |
| SetSystemInfoParameters        |        |
| GetSystemInfoParameters        |        |
| GetChannelAuthCapabilities     | √      |
| GetSessionChallenge            | √      |
| ActivateSession                | √      |
| SetSessionPrivilegeLevel       | √      |
| CloseSession                   | √      |
| GetSessionInfo                 | √      | session info                 |
| GetAuthCode                    | √      |
| SetChannelAccess               | √      | channel setaccess            |
| GetChannelAccess               | √      | channel info/getaccess       |
| GetChannelInfo                 | √      | channel info                 |
| SetUserAccess                  | √      |
| GetUserAccess                  | √      | user summary                 |
| GetUsers (*)                   | √      | user list                    |
| SetUsername                    | √      | user set name                |
| DisableUser (*)                | √      | user disable                 |
| EnableUser (*)                 | √      | user enable                  |
| GetUsername                    | √      |
| SetUserPassword                | √      | user set password            |
| TestUserPassword (*)           | √      | user test                    |
| ActivatePayload                |        |
| DeactivatePayload              |        |
| GetPayloadActivationStatus     |        |
| GetPayloadInstanceInfo         |        |
| SetUserPayloadAccess           |        |
| GetUserPayloadAccess           |        |
| GetChannelPayloadSupport       |        |
| GetChannelPayloadVersion       |        |
| GetChannelOEMPayloadInfo       |        |
| MasterWriteRead                |        |
| GetChannelCipherSuites         | √      |
| SuspendOrResumeEncryption      |        |
| SetChannelCipherSuites         |        |
| GetSystemInterfaceCapabilities | √      |

### Chassis Device Commands

| Method                    | Status | corresponding ipmitool usage                      |
| ------------------------- | ------ | ------------------------------------------------- |
| GetChassisCapabilities    | √      |
| GetChassisStatus          | √      | chassis status, chassis power status              |
| ChassisControl            | √      | chassis power on/off/cycle/reset/diag/soft        |
| ChassisReset              |        |
| ChassisIdentify           | √      | chassis identify                                  |
| SetChassisCapabilities    | √      |
| SetPowerRestorePolicy     | √      | chassis policy list/always-on/previous/always-off |
| GetSystemRestartCause     | √      | chassis restart_cause                             |
| SetSystemBootOptions      | √      | chassis bootparam set                             |
| SetBootParamBootFlags (*) | √      | chassis bootdev                                   |
| GetSystemBootOptions      | √      | chassis bootparam get                             |
| SetFrontPanelEnables      | √      |
| SetPowerCycleInterval     | √      |
| GetPOHCounter             | √      | chassis poh                                       |

### Event Commands

| Method               | Status | corresponding ipmitool usage |
| -------------------- | ------ | ---------------------------- |
| SetEventReceiver     | √      |
| GetEventReceiver     | √      |
| PlatformEventMessage | √      |

### PEF and Alerting Commands

| Method                  | Status | corresponding ipmitool usage |
| ----------------------- | ------ | ---------------------------- |
| GetPEFCapabilities      | √      | pef capabilities             |
| ArmPEFPostponeTimer     |        |
| SetPEFConfigParameters  |        |
| GetPEFConfigParameters  |        |
| SetLastProcessedEventId |        |
| GetLastProcessedEventId |        |
| AlertImmediate          |        |
| PEFAck                  |        |

### Sensor Device Commands

| Method                         | Status | corresponding ipmitool usage |
| ------------------------------ | ------ | ---------------------------- |
| GetDeviceSDRInfo               | √      |
| GetDeviceSDR                   | √      |
| ReserveDeviceSDRRepo           | √      |
| GetSensorReadingFactors        | √      |
| SetSensorHysteresis            | √      |
| GetSensorHysteresis            | √      |
| SetSensorThresholds            | √      |
| GetSensorThresholds            | √      |
| SetSensorEventEnable           |        |
| GetSensorEventEnable           | √      |
| RearmSensorEvents              |        |
| GetSensorEventStatus           | √      |
| GetSensorReading               | √      |
| SetSensorType                  | √      |
| GetSensorType                  | √      |
| SetSensorReadingAndEventStatus | √      |
| GetSensors (*)                 | √      | sensor list                  |
| GetSensorByID (*)              | √      |                              |
| GetSensorByName (*)            | √      | sensor get                   |

### FRU Device Commands

| Method                  | Status | corresponding ipmitool usage |
| ----------------------- | ------ | ---------------------------- |
| GetFRUInventoryAreaInfo | √      |
| ReadFRUData             | √      |
| WriteFRUData            | √      |


### SDR Device Commands

| Method                 | Status | corresponding ipmitool usage |
| ---------------------- | ------ | ---------------------------- |
| GetSDRRepoInfo         | √      | sdr info                     |
| GetSDRRepoAllocInfo    | √      | sdr info                     |
| ReserveSDRRepo         |        |
| GetSDR                 | √      |                              |
| GetSDRs (*)            | √      |                              |
| GetSDRBySensorID (*)   | √      |                              |
| GetSDRBySensorName (*) | √      |
| AddSDR                 |        |
| PartialAddSDR          |        |
| DeleteSDR              |        |
| ClearSDRRepo           |        |
| GetSDRRepoTime         |        |
| SetSDRRepoTime         |        |
| EnterSDRRepoUpateMode  |        |
| ExitSDRRepoUpdateMode  |        |
| RunInitializationAgent |        |

### SEL Device Commands

| Method              | Status | corresponding ipmitool usage |
| ------------------- | ------ | ---------------------------- |
| GetSELInfo          | √      | sel info                     |
| GetSELAllocInfo     | √      | sel info                     |
| ReserveSEL          | √      |
| GetSELEntry         | √      |
| AddSELEntry         | √      |
| PartialAddSELEntry  |        |
| DeleteSELEntry      | √      |
| ClearSEL            | √      | sel clear                    |
| GetSELTime          | √      |
| SetSELTime          | √      |
| GetAuxLogStatus     |        |
| SetAuxLogStatus     |        |
| GetSELTimeUTCOffset | √      |
| SetSELTimeUTCOffset | √      |

### LAN Device Commands

| Method             | Status | corresponding ipmitool usage |
| ------------------ | ------ | ---------------------------- |
| SetLanConfigParams |        |
| GetLanConfigParams | √      |
| SuspendARPs        | √      |
| GetIpStatistics    | √      |

### Serial/Modem Device Commands

| Method                 | Status | corresponding ipmitool usage |
| ---------------------- | ------ | ---------------------------- |
| SetSerialConfig        |        |
| GetSerialConfig        |        |
| SetSerialMux           |        |
| GetTapResponseCodes    |        |
| SetPPPTransmitData     |        |
| GetPPPTransmitData     |        |
| SendPPPPacket          |        |
| GetPPPReceiveData      |        |
| SerialConnectionActive |        |
| Callback               |        |
| SetUserCallbackOptions |        |
| GetUserCallbackOptions |        |
| SetSerialRoutingMux    |        |
| SOLActivating          | √      |
| GetSOLConfigParams     | √      |
| SetSOLConfigParams     | √      |
| SOLInfo                | √      | sol info                     |

### Command Forwarding Commands

| Method          | Status | corresponding ipmitool usage |
| --------------- | ------ | ---------------------------- |
| Fowarded        |        |
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
