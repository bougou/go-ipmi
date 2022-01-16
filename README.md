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

### IPM Device Global Commands

| Method                             | Status | corresponding ipmitool usage |
| ---------------------------------- | ------ | ---------------------------- |
| GetDeviceID                        | √      | mc info                      |
| ColdReset                          | √      | mc reset cold                |
| WarmReset                          | √      | mc reset warm                |
| GetSelfTestResults                 | √      | mc selftest                  |
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
| ResetWatchdogTimer |        |
| SetWatchdogTimer   |        |
| GetWatchdogTimer   |        |

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
| GetSessionInfo                 | √      |
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
| TestUserPassword(*)            | √      | user test                    |
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

| Method                 | Status | corresponding ipmitool usage                 |
| ---------------------- | ------ | -------------------------------------------- |
| GetChassisCapabilities | √      |
| GetChassisStatus       | √      | chassis status                               |
| ChassisControl         | √      | chassis power on/off/cycle/reset/diag/soft   |
| ChassisReset           |        |
| ChassisIdentify        | √      | chassis identify                             |
| SetChassisCapabilities | √      |
| SetPowerRestorePolicy  | √      | chassis policy always-on/previous/always-off |
| GetSystemRestartCause  | √      | chassis restart_cause                        |
| SetSystemBootOptions   | √      |
| GetSystemBootOptions   | √      |
| SetFrontPanelEnables   | √      |
| SetPowerCycleInterval  | √      |
| GetPOHCounter          |        |

### Event Commands

| Method           | Status | corresponding ipmitool usage |
| ---------------- | ------ | ---------------------------- |
| SetEventReceiver |        |
| GetEventReceiver |        |
| EventMessage     |        |

### PEF and Alerting Commands

| Method                  | Status | corresponding ipmitool usage |
| ----------------------- | ------ | ---------------------------- |
| GetPefCapabilities      |        |
| ArmPefPostponeTimer     |        |
| SetPefConfigParameters  |        |
| GetPefConfigParameters  |        |
| SetLastProcessedEventId |        |
| GetLastProcessedEventId |        |
| AlertImmediate          |        |
| PetAck                  |        |

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
| GetSDR                 | √      | sdr get                      |
| GetSDRs (*)            | √      | sdr list/elist               |
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
| AddSELEntry         |        |
| PartialAddSELEntry  |        |
| DeleteSELEntry      |        |
| ClearSEL            | √      | sel clear                    |
| GetSELTime          |        |
| SetSELTime          |        |
| GetAuxLogStatus     |        |
| SetAuxLogStatus     |        |
| GetSELTimeUtcOffset |        |
| SetSELTimeUtcOffset |        |

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
| SolActivating          |        |
| GetSolConfigParams     |        |
| SetSolConfigParams     |        |

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
