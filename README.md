<meta name="author" content="Bougou">
<meta name="description" content="Go IPMI library">
<meta name="keywords" content="ipmi, go, golang, bmc">

# [go-ipmi](https://github.com/bougou/go-ipmi)

[`go-ipmi`](https://github.com/bougou/go-ipmi) is a pure Golang native IPMI library. It DOES NOT wrap `ipmitool`.

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
	// Supports local mode client when running directly on Linux
	// client, err := ipmi.NewOpenClient()
	if err != nil {
		panic(err)
	}

	// You can optionally enable debug mode
	// client.WithDebug(true)

	// You can set the interface type. Valid options are: open/lan/lanplus/tool (default: open)
	// client.WithInterface(ipmi.InterfaceLanplus)

	// !!! Note !!!
	// From v0.6.0, all IPMI command methods of the Client require a context as the first argument.
	ctx := context.Background()

	// Connect creates an authenticated session
	if err := client.Connect(ctx); err != nil {
		panic(err)
	}

	// Now you can execute other IPMI commands that require authentication

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

The `goipmi` binary provides command usage similar to `ipmitool`.
The `goipmi` tool uses the `go-ipmi` library under the hood.

The purpose of creating the `goipmi` tool was not to substitute `ipmitool`.
It was created to verify the correctness of the `go-ipmi` library.

## Functions Comparison with ipmitool

Each command defined in the IPMI specification consists of a pair of request/response messages.
These IPMI commands are implemented as methods of the `ipmi.Client` struct in this library.

Some `ipmitool` command line operations are implemented by calling just one IPMI command,
while others are not. For example, `ipmitool sdr list` involves a loop of `GetSDR` IPMI commands.

This library also implements some helper methods that are not IPMI commands defined
in the IPMI specification, but are common utilities, like `GetSDRs` to get all SDRs.
These methods are marked with an asterisk (*) after the method name in the following documentation.

The implementation logic of IPMI commands is largely consistent. See [Contributing](./CONTRIBUTING.md)

> More commands are in development...

### IPM Device Global Commands


| Method                             | Status             | corresponding ipmitool usage  |
| ---------------------------------- | ------------------ | ----------------------------- |
| GetDeviceID                        | :white_check_mark: | mc info                       |
| ColdReset                          | :white_check_mark: | mc reset cold                 |
| WarmReset                          | :white_check_mark: | mc reset warm                 |
| GetSelfTestResults                 | :white_check_mark: | mc selftest, chassis selftest |
| ManufacturingTestOn                | :white_check_mark: |                               |
| SetACPIPowerState                  | :white_check_mark: |                               |
| GetACPIPowerState                  | :white_check_mark: |                               |
| GetDeviceGUID                      | :white_check_mark: |                               |
| GetNetFnSupport                    | :white_check_mark: |                               |
| GetCommandSupport                  | :white_check_mark: |                               |
| GetCommandSubfunctionSupport       | :white_check_mark: |                               |
| GetConfigurableCommands            | :white_check_mark: |                               |
| GetConfigurableCommandSubfunctions | :white_check_mark: |                               |
| SetCommandEnables                  | :white_check_mark: |                               |
| GetCommandEnables                  | :white_check_mark: |                               |
| SetCommandSubfunctionEnables       | :white_check_mark: |                               |
| GetCommandSubfunctionEnables       | :white_check_mark: |                               |
| GetOEMNetFnIanaSupport             |                    |                               |

### BMC Watchdog Timer Commands

| Method             | Status             | corresponding ipmitool usage |
| ------------------ | ------------------ | ---------------------------- |
| ResetWatchdogTimer | :white_check_mark: | mc watchdog reset            |
| SetWatchdogTimer   | :white_check_mark: |                              |
| GetWatchdogTimer   | :white_check_mark: | mc watchdog get              |

### BMC Device and Messaging Commands

| Method                         | Status             | corresponding ipmitool usage |
| ------------------------------ | ------------------ | ---------------------------- |
| SetBMCGlobalEnables            | :white_check_mark: |                              |
| GetBMCGlobalEnables            | :white_check_mark: |                              |
| ClearMessageFlags              | :white_check_mark: |                              |
| GetMessageFlags                | :white_check_mark: |                              |
| EnableMessageChannelReceive    | :white_check_mark: |                              |
| GetMessage                     | :white_check_mark: |                              |
| SendMessage                    | :white_check_mark: |                              |
| ReadEventMessageBuffer         | :white_check_mark: |                              |
| GetBTInterfaceCapabilities     | :white_check_mark: |                              |
| GetSystemGUID                  | :white_check_mark: | mc guid                      |
| SetSystemInfoParam             | :white_check_mark: |                              |
| SetSystemInfoParamFor (*)      | :white_check_mark: |                              |
| GetSystemInfoParam             | :white_check_mark: |                              |
| GetSystemInfoParamFor (*)      | :white_check_mark: |                              |
| GetSystemInfoParams (*)        | :white_check_mark: |                              |
| GetSystemInfoParamsFor (*)     | :white_check_mark: |                              |
| GetSystemInfo (*)              | :white_check_mark: |                              |
| GetChannelAuthCapabilities     | :white_check_mark: |                              |
| GetSessionChallenge            | :white_check_mark: |                              |
| ActivateSession                | :white_check_mark: |                              |
| SetSessionPrivilegeLevel       | :white_check_mark: |                              |
| CloseSession                   | :white_check_mark: |                              |
| GetSessionInfo                 | :white_check_mark: | session info                 |
| GetAuthCode                    | :white_check_mark: |                              |
| SetChannelAccess               | :white_check_mark: | channel setaccess            |
| GetChannelAccess               | :white_check_mark: | channel info/getaccess       |
| GetChannelInfo                 | :white_check_mark: | channel info                 |
| SetUserAccess                  | :white_check_mark: |                              |
| GetUserAccess                  | :white_check_mark: | user summary                 |
| GetUsers (*)                   | :white_check_mark: | user list                    |
| SetUsername                    | :white_check_mark: | user set name                |
| DisableUser (*)                | :white_check_mark: | user disable                 |
| EnableUser (*)                 | :white_check_mark: | user enable                  |
| GetUsername                    | :white_check_mark: |
| SetUserPassword                | :white_check_mark: | user set password            |
| TestUserPassword (*)           | :white_check_mark: | user test                    |
| ActivatePayload                | :white_check_mark: |                              |
| DeactivatePayload              | :white_check_mark: |                              |
| GetPayloadActivationStatus     | :white_check_mark: |                              |
| GetPayloadInstanceInfo         | :white_check_mark: |                              |
| SetUserPayloadAccess           | :white_check_mark: |                              |
| GetUserPayloadAccess           | :white_check_mark: | sol payload status           |
| GetChannelPayloadSupport       | :white_check_mark: |                              |
| GetChannelPayloadVersion       | :white_check_mark: |                              |
| GetChannelOEMPayloadInfo       | :white_check_mark: |                              |
| MasterWriteRead                | :white_check_mark: |                              |
| GetChannelCipherSuites         | :white_check_mark: |                              |
| SuspendResumePayloadEncryption | :white_check_mark: |                              |
| SetChannelSecurityKeys         | :white_check_mark: |                              |
| GetSystemInterfaceCapabilities | :white_check_mark: |                              |

### Chassis Device Commands

| Method                            | Status             | corresponding ipmitool usage                      |
| --------------------------------- | ------------------ | ------------------------------------------------- |
| GetChassisCapabilities            | :white_check_mark: |                                                   |
| GetChassisStatus                  | :white_check_mark: | chassis status, chassis power status              |
| ChassisControl                    | :white_check_mark: | chassis power on/off/cycle/reset/diag/soft        |
| ChassisReset                      | :white_check_mark: |                                                   |
| ChassisIdentify                   | :white_check_mark: | chassis identify                                  |
| SetChassisCapabilities            | :white_check_mark: |                                                   |
| SetPowerRestorePolicy             | :white_check_mark: | chassis policy list/always-on/previous/always-off |
| GetSystemRestartCause             | :white_check_mark: | chassis restart_cause                             |
| SetBootParamBootFlags (*)         | :white_check_mark: | chassis bootdev                                   |
| SetBootDevice (*)                 | :white_check_mark: | chassis bootdev                                   |
| SetSystemBootOptionsParam         | :white_check_mark: | chassis bootparam set                             |
| GetSystemBootOptionsParam         | :white_check_mark: | chassis bootparam get                             |
| GetSystemBootOptionsParamFor (*)  | :white_check_mark: | chassis bootparam get                             |
| GetSystemBootOptionsParams (*)    | :white_check_mark: | chassis bootparam get                             |
| GetSystemBootOptionsParamsFor (*) | :white_check_mark: | chassis bootparam get                             |
| SetFrontPanelEnables              | :white_check_mark: |                                                   |
| SetPowerCycleInterval             | :white_check_mark: |                                                   |
| GetPOHCounter                     | :white_check_mark: | chassis poh                                       |

### Event Commands

| Method               | Status             | corresponding ipmitool usage |
| -------------------- | ------------------ | ---------------------------- |
| SetEventReceiver     | :white_check_mark: |                              |
| GetEventReceiver     | :white_check_mark: |                              |
| PlatformEventMessage | :white_check_mark: |                              |

### PEF and Alerting Commands

| Method                    | Status             | corresponding ipmitool usage |
| ------------------------- | ------------------ | ---------------------------- |
| GetPEFCapabilities        | :white_check_mark: | pef capabilities             |
| ArmPEFPostponeTimer       | :white_check_mark: |                              |
| SetPEFConfigParam         | :white_check_mark: |                              |
| GetPEFConfigParam         | :white_check_mark: |                              |
| GetPEFConfigParamFor (*)  | :white_check_mark: |                              |
| GetPEFConfigParams (*)    | :white_check_mark: |                              |
| GetPEFConfigParamsFor (*) | :white_check_mark: |                              |
| SetLastProcessedEventId   | :white_check_mark: |                              |
| GetLastProcessedEventId   | :white_check_mark: |                              |
| AlertImmediate            | :white_check_mark: |                              |
| PETAcknowledge            | :white_check_mark: |                              |

### Sensor Device Commands

| Method                         | Status             | corresponding ipmitool usage |
| ------------------------------ | ------------------ | ---------------------------- |
| GetDeviceSDRInfo               | :white_check_mark: |                              |
| GetDeviceSDR                   | :white_check_mark: |                              |
| ReserveDeviceSDRRepo           | :white_check_mark: |                              |
| GetSensorReadingFactors        | :white_check_mark: |                              |
| SetSensorHysteresis            | :white_check_mark: |                              |
| GetSensorHysteresis            | :white_check_mark: |                              |
| SetSensorThresholds            | :white_check_mark: |                              |
| GetSensorThresholds            | :white_check_mark: |                              |
| SetSensorEventEnable           | :white_check_mark: |                              |
| GetSensorEventEnable           | :white_check_mark: |                              |
| RearmSensorEvents              | :white_check_mark: |                              |
| GetSensorEventStatus           | :white_check_mark: |                              |
| GetSensorReading               | :white_check_mark: |                              |
| SetSensorType                  | :white_check_mark: |                              |
| GetSensorType                  | :white_check_mark: |                              |
| SetSensorReadingAndEventStatus | :white_check_mark: |                              |
| GetSensors (*)                 | :white_check_mark: | sensor list, sdr type        |
| GetSensorByID (*)              | :white_check_mark: |                              |
| GetSensorByName (*)            | :white_check_mark: | sensor get                   |

### FRU Device Commands

| Method                  | Status             | corresponding ipmitool usage |
| ----------------------- | ------------------ | ---------------------------- |
| GetFRUInventoryAreaInfo | :white_check_mark: |                              |
| ReadFRUData             | :white_check_mark: |                              |
| WriteFRUData            | :white_check_mark: |                              |
| GetFRU (*)              | :white_check_mark: | fru print                    |
| GetFRUs (*)             | :white_check_mark: | fru print                    |


### SDR Device Commands

| Method                 | Status             | corresponding ipmitool usage |
| ---------------------- | ------------------ | ---------------------------- |
| GetSDRRepoInfo         | :white_check_mark: | sdr info                     |
| GetSDRRepoAllocInfo    | :white_check_mark: | sdr info                     |
| ReserveSDRRepo         |                    |                              |
| GetSDR                 | :white_check_mark: |                              |
| GetSDRs (*)            | :white_check_mark: |                              |
| GetSDRBySensorID (*)   | :white_check_mark: |                              |
| GetSDRBySensorName (*) | :white_check_mark: |                              |
| AddSDR                 |                    |                              |
| PartialAddSDR          |                    |                              |
| DeleteSDR              |                    |                              |
| ClearSDRRepo           |                    |                              |
| GetSDRRepoTime         |                    |                              |
| SetSDRRepoTime         |                    |                              |
| EnterSDRRepoUpdateMode |                    |                              |
| ExitSDRRepoUpdateMode  |                    |                              |
| RunInitializationAgent |                    |                              |

### SEL Device Commands

| Method              | Status             | corresponding ipmitool usage |
| ------------------- | ------------------ | ---------------------------- |
| GetSELInfo          | :white_check_mark: | sel info                     |
| GetSELAllocInfo     | :white_check_mark: | sel info                     |
| ReserveSEL          | :white_check_mark: |                              |
| GetSELEntry         | :white_check_mark: |                              |
| AddSELEntry         | :white_check_mark: |                              |
| PartialAddSELEntry  |                    |                              |
| DeleteSELEntry      | :white_check_mark: |                              |
| ClearSEL            | :white_check_mark: | sel clear                    |
| GetSELTime          | :white_check_mark: |                              |
| SetSELTime          | :white_check_mark: |                              |
| GetAuxLogStatus     |                    |                              |
| SetAuxLogStatus     |                    |                              |
| GetSELTimeUTCOffset | :white_check_mark: |                              |
| SetSELTimeUTCOffset | :white_check_mark: |                              |

### LAN Device Commands

| Method                    | Status             | corresponding ipmitool usage |
| ------------------------- | ------------------ | ---------------------------- |
| SetLanConfigParam         | :white_check_mark: | lan set                      |
| SetLanConfigParamFor (*)  | :white_check_mark: | lan set                      |
| GetLanConfigParam         | :white_check_mark: |                              |
| GetLanConfigParamFor (*)  | :white_check_mark: | lan print                    |
| GetLanConfigParams (*)    | :white_check_mark: | lan print                    |
| GetLanConfigParamsFor (*) | :white_check_mark: | lan print                    |
| GetLanConfig (*)          | :white_check_mark: | lan print                    |
| SuspendARPs               | :white_check_mark: |                              |
| GetIPStatistics           | :white_check_mark: |                              |

### Serial/Modem Device Commands

| Method                    | Status             | corresponding ipmitool usage |
| ------------------------- | ------------------ | ---------------------------- |
| SetSerialConfig           |                    |                              |
| GetSerialConfig           |                    |                              |
| SetSerialMux              |                    |                              |
| GetTapResponseCodes       |                    |                              |
| SetPPPTransmitData        |                    |                              |
| GetPPPTransmitData        |                    |                              |
| SendPPPPacket             |                    |                              |
| GetPPPReceiveData         |                    |                              |
| SerialConnectionActive    |                    |                              |
| Callback                  |                    |                              |
| SetUserCallbackOptions    |                    |                              |
| GetUserCallbackOptions    |                    |                              |
| SetSerialRoutingMux       |                    |                              |
| SOLActivating             | :white_check_mark: |                              |
| SetSOLConfigParam         | :white_check_mark: |                              |
| SetSOLConfigParamFor (*)  | :white_check_mark: |                              |
| GetSOLConfigParam         | :white_check_mark: |                              |
| GetSOLConfigParamFor (*)  | :white_check_mark: |                              |
| GetSOLConfigParams (*)    | :white_check_mark: | sol info                     |
| GetSOLConfigParamsFor (*) | :white_check_mark: | sol info                     |

### Command Forwarding Commands

| Method          | Status | corresponding ipmitool usage |
| --------------- | ------ | ---------------------------- |
| Forwarded       |        |                              |
| SetForwarded    |        |                              |
| GetForwarded    |        |                              |
| EnableForwarded |        |                              |

### Bridge Management Commands (ICMB)

| Method                | Status | corresponding ipmitool usage |
| --------------------- | ------ | ---------------------------- |
| GetBridgeState        |        |                              |
| SetBridgeState        |        |                              |
| GetICMBAddress        |        |                              |
| SetICMBAddress        |        |                              |
| SetBridgeProxyAddress |        |                              |
| GetBridgeStatistics   |        |                              |
| GetICMBCapabilities   |        |                              |
| ClearBridgeStatistics |        |                              |
| GetBridgeProxyAddress |        |                              |
| GetICMBConnectorInfo  |        |                              |
| GetICMBConnectionID   |        |                              |
| SendICMBConnectionID  |        |                              |

### Discovery Commands (ICMB)

| Method              | Status | corresponding ipmitool usage |
| ------------------- | ------ | ---------------------------- |
| PrepareForDiscovery |        |                              |
| GetAddresses        |        |                              |
| SetDiscovered       |        |                              |
| GetChassisDeviceId  |        |                              |
| SetChassisDeviceId  |        |                              |

### Bridging Commands (ICMB)

| Method        | Status | corresponding ipmitool usage |
| ------------- | ------ | ---------------------------- |
| BridgeRequest |        |                              |
| BridgeMessage |        |                              |

### Event Commands (ICMB)

| Method                 | Status | corresponding ipmitool usage |
| ---------------------- | ------ | ---------------------------- |
| GetEventCount          |        |                              |
| SetEventDestination    |        |                              |
| SetEventReceptionState |        |                              |
| SendICMBEventMessage   |        |                              |
| GetEventDestination    |        |                              |
| GetEventReceptionState |        |                              |


### Other Bridge Commands

| Method      | Status | corresponding ipmitool usage |
| ----------- | ------ | ---------------------------- |
| ErrorReport |        |                              |

### DCMI Commands

| Method                          | Status             | corresponding ipmitool usage |
| ------------------------------- | ------------------ | ---------------------------- |
| GetDCMICapParam                 | :white_check_mark: | dcmi discovery               |
| GetDCMICapParamFor (*)          | :white_check_mark: | dcmi discovery               |
| GetDCMICapParams (*)            | :white_check_mark: | dcmi discovery               |
| GetDCMICapParamsFor (*)         | :white_check_mark: | dcmi discovery               |
| GetDCMIPowerReading             | :white_check_mark: | dcmi power reading           |
| GetDCMIPowerLimit               | :white_check_mark: | dcmi power get_limit         |
| SetDCMIPowerLimit               | :white_check_mark: | dcmi power set_limit         |
| ActivateDCMIPowerLimit          | :white_check_mark: | dcmi activate/deactivate     |
| GetDCMIAssetTag                 | :white_check_mark: | dcmi asset_tag               |
| GetDCMIAssetTagFull (*)         | :white_check_mark: | dcmi asset_tag               |
| GetDCMISensorInfo               | :white_check_mark: | dcmi sensors                 |
| SetDCMIAssetTag                 | :white_check_mark: | dcmi set_asset_tag           |
| GetDCMIMgmtControllerIdentifier | :white_check_mark: | dcmi get_mc_id_string        |
| SetDCMIMgmtControllerIdentifier | :white_check_mark: | dcmi set_mc_id_string        |
| SetDCMIThermalLimit             | :white_check_mark: | dcmi thermalpolicy get       |
| GetDCMIThermalLimit             | :white_check_mark: | dcmi thermalpolicy set       |
| GetDCMITemperatureReadings      | :white_check_mark: | dcmi get_temp_reading        |
| SetDCMIConfigParam              | :white_check_mark: | dcmi set_conf_param          |
| GetDCMIConfigParam              | :white_check_mark: | dcmi get_conf_param          |
| GetDCMIConfigParamFor (*)       | :white_check_mark: | dcmi get_conf_param          |
| GetDCMIConfigParams (*)         | :white_check_mark: | dcmi get_conf_param          |
| GetDCMIConfigParamsFor (*)      | :white_check_mark: | dcmi get_conf_param          |

## Reference

- [Intelligent Platform Management Interface Specification Second Generation v2.0](https://www.intel.com/content/dam/www/public/us/en/documents/specification-updates/ipmi-intelligent-platform-mgt-interface-spec-2nd-gen-v2-0-spec-update.pdf)
- [Platform Management FRU Information Storage Definition](https://www.intel.com/content/dam/www/public/us/en/documents/specification-updates/ipmi-platform-mgt-fru-info-storage-def-v1-0-rev-1-3-spec-update.pdf)
- [PC SDRAM Serial Presence Detect (SPD) Specification](https://cdn.hackaday.io/files/10119432931296/Spdsd12b.pdf)
- [DCMI Group Extension Specification v1.5](https://www.intel.com/content/dam/www/public/us/en/documents/technical-specifications/dcmi-v1-5-rev-spec.pdf)
