package commands

import (
	"fmt"

	"github.com/bougou/go-ipmi"
	"github.com/spf13/cobra"
)

func NewCmdSensor() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sensor",
		Short: "sensor",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initClient()
		},
		Run: func(cmd *cobra.Command, args []string) {
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return closeClient()
		},
	}
	cmd.AddCommand(NewCmdSensorDeviceSDRInfo())
	cmd.AddCommand(NewCmdSensorGet())
	cmd.AddCommand(NewCmdSensorList())
	cmd.AddCommand(NewCmdSensorThreshold())
	cmd.AddCommand(NewCmdSensorEventEnable())
	cmd.AddCommand(NewCmdSensorEventStatus())
	cmd.AddCommand(NewCmdSensorReading())
	cmd.AddCommand(NewCmdSensorReadingFactors())
	cmd.AddCommand(NewCmdSensorDetail())

	return cmd
}

func NewCmdSensorDeviceSDRInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "info",
		Run: func(cmd *cobra.Command, args []string) {
			res, err := client.GetDeviceSDRInfo(true)
			if err != nil {
				CheckErr(fmt.Errorf("GetDeviceSDRInfo failed, err: %s", err))
			}
			fmt.Println(res.Format())
		},
	}
	return cmd
}

func NewCmdSensorList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list",
		Run: func(cmd *cobra.Command, args []string) {

			sensors, err := client.GetSensors()
			if err != nil {
				CheckErr(fmt.Errorf("GetSensors failed, err: %s", err))
			}

			fmt.Println(ipmi.FormatSensors(false, sensors...))
		},
	}

	return cmd
}

func NewCmdSensorGet() *cobra.Command {
	usage := `sensor get <sensorNumber> or <sensorName>, sensorName should be quoted if contains space`

	cmd := &cobra.Command{
		Use:   "get",
		Short: "get",
		Run: func(cmd *cobra.Command, args []string) {
			var sensorNumber uint8

			if len(args) < 1 {
				CheckErr(fmt.Errorf("no Sensor ID or Sensor Name supplied, usage: %s", usage))
			}

			var sensor *ipmi.Sensor
			var err error

			id, err := parseStringToInt64(args[0])
			if err != nil {
				// suppose args is sensor name
				sensor, err = client.GetSensorByName(args[0])
				if err != nil {
					CheckErr(fmt.Errorf("GetSensorByName failed, err: %s", err))
				}
			} else {
				sensorNumber = uint8(id)
				sensor, err = client.GetSensorByID(sensorNumber)
				if err != nil {
					CheckErr(fmt.Errorf("GetSensorByID failed, err: %s", err))
				}
			}

			client.Debug("sensor", sensor)
			fmt.Println(sensor)
		},
	}
	return cmd
}

func NewCmdSensorThreshold() *cobra.Command {
	usage := `
sensor threshold get <sensor_number>
	`
	cmd := &cobra.Command{
		Use:   "threshold",
		Short: "threshold",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 2 {
				CheckErr(fmt.Errorf("usage: %s", usage))
			}

			action := args[0]

			var sensorNumber uint8
			i, err := parseStringToInt64(args[1])
			if err != nil {
				CheckErr(fmt.Errorf("invalid sensor number, err: %s", err))
			}
			sensorNumber = uint8(i)

			switch action {
			case "get":
				res, err := client.GetSensorThresholds(sensorNumber)
				if err != nil {
					CheckErr(fmt.Errorf("GetSensorThresholds failed, err: %s", err))
				}
				fmt.Println(res.Format())
			case "set":
			default:
				CheckErr(fmt.Errorf("usage: %s", usage))
			}

		},
	}
	return cmd
}

func NewCmdSensorEventStatus() *cobra.Command {
	usage := `
sensor event-status get <sensor_number>
	`
	cmd := &cobra.Command{
		Use:   "event-status ",
		Short: "event-status ",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 2 {
				CheckErr(fmt.Errorf("usage: %s", usage))
			}

			action := args[0]

			var sensorNumber uint8
			i, err := parseStringToInt64(args[1])
			if err != nil {
				CheckErr(fmt.Errorf("invalid sensor number, err: %s", err))
			}
			sensorNumber = uint8(i)

			switch action {
			case "get":
				res, err := client.GetSensorEventStatus(sensorNumber)
				if err != nil {
					CheckErr(fmt.Errorf("GetSensorEventStatus failed, err: %s", err))
				}
				fmt.Println(res.Format())
			case "set":
			default:
				CheckErr(fmt.Errorf("usage: %s", usage))
			}
		},
	}
	return cmd
}

func NewCmdSensorEventEnable() *cobra.Command {
	usage := `
sensor event-enable get <sensor_number>
	`
	cmd := &cobra.Command{
		Use:   "event-enable ",
		Short: "event-enable ",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 2 {
				CheckErr(fmt.Errorf("usage: %s", usage))
			}

			action := args[0]

			var sensorNumber uint8
			i, err := parseStringToInt64(args[1])
			if err != nil {
				CheckErr(fmt.Errorf("invalid sensor number, err: %s", err))
			}
			sensorNumber = uint8(i)

			switch action {
			case "get":
				res, err := client.GetSensorEventEnable(sensorNumber)
				if err != nil {
					CheckErr(fmt.Errorf("GetSensorEventEnable failed, err: %s", err))
				}
				fmt.Println(res.Format())
			case "set":
			default:
				CheckErr(fmt.Errorf("usage: %s", usage))
			}
		},
	}
	return cmd
}

func NewCmdSensorReading() *cobra.Command {
	usage := `
sensor reading get <sensor_number>
	`
	cmd := &cobra.Command{
		Use:   "reading",
		Short: "reading",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 2 {
				CheckErr(fmt.Errorf("usage: %s", usage))
			}

			action := args[0]

			var sensorNumber uint8
			i, err := parseStringToInt64(args[1])
			if err != nil {
				CheckErr(fmt.Errorf("invalid sensor number, err: %s", err))
			}
			sensorNumber = uint8(i)

			switch action {
			case "get":
				res, err := client.GetSensorReading(sensorNumber)
				if err != nil {
					CheckErr(fmt.Errorf("GetSensorReading failed, err: %s", err))
				}
				fmt.Println(res.Format())
			case "set":
			default:
				CheckErr(fmt.Errorf("usage: %s", usage))
			}
		},
	}
	return cmd
}

func NewCmdSensorReadingFactors() *cobra.Command {
	usage := `
sensor reading-factors get <sensor_number>
	`
	cmd := &cobra.Command{
		Use:   "reading-factors",
		Short: "reading-factors",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 2 {
				CheckErr(fmt.Errorf("usage: %s", usage))
			}

			action := args[0]

			var sensorNumber uint8
			i, err := parseStringToInt64(args[1])
			if err != nil {
				CheckErr(fmt.Errorf("invalid sensor number, err: %s", err))
			}
			sensorNumber = uint8(i)

			switch action {
			case "get":
				res0, err := client.GetSensorReading(sensorNumber)
				if err != nil {
					CheckErr(fmt.Errorf("GetSensorReading failed, err: %s", err))
				}
				fmt.Println(res0.Format())

				res, err := client.GetSensorReadingFactors(sensorNumber, res0.AnalogReading)
				if err != nil {
					CheckErr(fmt.Errorf("GetSensorReadingFactors failed, err: %s", err))
				}
				fmt.Println(res.Format())
			case "set":
			default:
				CheckErr(fmt.Errorf("usage: %s", usage))
			}
		},
	}
	return cmd
}

func NewCmdSensorDetail() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "detail",
		Short: "detail",
		Run: func(cmd *cobra.Command, args []string) {
			var sensorNumber uint8

			if len(args) >= 1 {
				i, err := parseStringToInt64(args[0])
				if err != nil {
					CheckErr(fmt.Errorf("invalid sensor number, err: %s", err))
				}
				sensorNumber = uint8(i)
			}

			sensor, err := client.GetSensorByID(sensorNumber)
			if err != nil {
				if err != nil {
					CheckErr(fmt.Errorf("GetSensorByID failed, err: %s", err))
				}
			}
			fmt.Println(sensor)
		},
	}
	return cmd
}
