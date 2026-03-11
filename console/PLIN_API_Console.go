package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"strconv" 
	"unsafe"
	"time"
	"runtime"
	plin "github.com/hail33scomet/goplin"
)

type ConsoleApp struct {
	Client     plin.HLINCLIENT
	Hardware   plin.HLINHW
	HwMode     plin.TLINHardwareMode
	HwBaudrate plin.WORD
	Mask       plin.UINT64
	PIDs       map[plin.BYTE]int
}

type HardwareInfo struct {
	Handle  plin.HLINHW
	Type    string
	Device  int
	Channel int
}

func NewConsoleApp() (*ConsoleApp, error) {

	app := &ConsoleApp{
		PIDs: make(map[plin.BYTE]int),
	}

	err := app.initialize()
	if err != nil {
		return nil, err
	}

	return app, nil
}

func (app *ConsoleApp) initialize() error {

	app.Client = 0
	app.Hardware = 0
	app.HwMode = plin.TLIN_HARDWAREMODE_NONE
	app.HwBaudrate = 0
	app.Mask = 0xFFFFFFFFFFFFFFFF

	// Initialize PID mapping
	for i := 0; i < 64; i++ {
		var pid plin.BYTE

		err := plin.LIN_GetPID(&pid)
		if err != nil {
			return err
		}
		app.PIDs[pid] = i
	}

	return nil
}

func (app *ConsoleApp) uninitialize() error {

	if app.Client != 0 {

		err := app.doLinDisconnect()
		if err != nil {
			return err
		}

		err = plin.LIN_RemoveClient(&app.Client)
		if err != nil {
			return err
		}

		app.Client = 0
	}

	return nil
}

func (app *ConsoleApp) doLinConnect(
	hw plin.HLINHW,
	mode plin.TLINHardwareMode,
	baud plin.WORD,
) error {

	// Disconnect existing hardware
	if app.Hardware != 0 {
		if err := app.doLinDisconnect(); err != nil {
			return fmt.Errorf("failed to disconnect existing hardware: %w", err)
		}
	}

	// Register client
	if app.Client == 0 {
		if err := plin.LIN_RegisterClient(
			"PLIN-API Console",
			0,
			&app.Client,
		); err != nil {
			return fmt.Errorf("failed to register client: %w", err)
		}
	}

	// Connect client
	if err := plin.LIN_ConnectClient(&app.Client, hw); err != nil {
		return fmt.Errorf("failed to connect client to hardware: %w", err)
	}

	app.Hardware = hw

	// Read hardware parameters
	var currMode plin.TLINHardwareMode
	if err := plin.LIN_GetHardwareParam(
		hw,
		plin.TLIN_HARDWAREPARAM_MODE,
		(*plin.BYTE)(unsafe.Pointer(&currMode)),
		plin.WORD(unsafe.Sizeof(currMode)),
	); err != nil {
		return fmt.Errorf("failed to get hardware mode: %w", err)
	}

	var currBaud plin.WORD
	if err := plin.LIN_GetHardwareParam(
		hw,
		plin.TLIN_HARDWAREPARAM_BAUDRATE,
		(*plin.BYTE)(unsafe.Pointer(&currBaud)),
		plin.WORD(unsafe.Sizeof(currBaud)),
	); err != nil {
		return fmt.Errorf("failed to get hardware baudrate: %w", err)
	}

	// Initialize if needed
	if currMode == plin.TLIN_HARDWAREMODE_NONE || currBaud != baud {
		if err := plin.LIN_InitializeHardware(
			&app.Client,
			app.Hardware,
			plin.WORD(mode),
			baud,
		); err != nil {
			app.Hardware = 0
			return fmt.Errorf("failed to initialize hardware: %w", err)
		}
	}

	app.HwMode = mode
	app.HwBaudrate = baud

	// Set filter
	if err := plin.LIN_SetClientFilter(
		app.Client,
		app.Hardware,
		&app.Mask,
	); err != nil {
		return fmt.Errorf("failed to set client filter: %w", err)
	}

	app.readFrameTableFromHw()

	return nil
}

func (app *ConsoleApp) doLinDisconnect() error {

	if app.Hardware == 0 {
		return nil
	}

	var clients [255]plin.HLINCLIENT

	err := plin.LIN_GetHardwareParam(
		app.Hardware,
		plin.TLIN_HARDWAREPARAM_CONNECTED_CLIENTS,
		(*plin.BYTE)(unsafe.Pointer(&clients[0])),
		plin.WORD(unsafe.Sizeof(clients)),
	)
	if err != nil {
		return fmt.Errorf("failed to get connected clients: %w", err)
	}

	lfOtherClient := false
	lfOwnClient := false

	for _, c := range clients {
		if c == 0 {
			continue
		}
		if c != app.Client {
			lfOtherClient = true
		}
		if c == app.Client {
			lfOwnClient = true
		}
	}

	if !lfOtherClient {
		if err := plin.LIN_ResetHardwareConfig(&app.Client, app.Hardware); err != nil {
			return fmt.Errorf("failed to reset hardware config: %w", err)
		}
	}

	if lfOwnClient {
		if err := plin.LIN_DisconnectClient(app.Client, app.Hardware); err != nil {
			return fmt.Errorf("failed to disconnect client: %w", err)
		}
		app.Hardware = 0
	}

	return nil
}

func (app *ConsoleApp) readFrameTableFromHw() ([]plin.TLINFrameEntry, error) {

	var result []plin.TLINFrameEntry
	app.Mask = 0

	for i := 0; i < 64; i++ {
		var frame plin.TLINFrameEntry
		frame.FrameId = plin.BYTE(i)
		frame.ChecksumType = plin.TLIN_CHECKSUMTYPE_AUTO
		frame.Direction = plin.TLIN_DIRECTION_SUBSCRIBER_AUTOLENGTH

		switch {
		case i <= 0x1F:
			frame.Length = 2
		case i <= 0x2F:
			frame.Length = 4
		default:
			frame.Length = 8
		}

		err := plin.LIN_GetFrameEntry(app.Hardware, &frame)
		if err != nil {
			continue
		}

		result = append(result, frame)

		if frame.Direction != plin.TLIN_DIRECTION_DISABLED {
			maskBit := plin.UINT64(1) << i
			app.Mask |= maskBit
		}

		if app.Client != 0 && app.Hardware != 0 {
			_ = plin.LIN_SetClientFilter(app.Client, app.Hardware, &app.Mask)
		}
	}

	return result, nil
}

func (app *ConsoleApp) getAvailableHardware() ([]HardwareInfo, error) {

	var result []HardwareInfo

	var count plin.WORD
	buf := make([]plin.HLINHW, 16)

	err := plin.LIN_GetAvailableHardware(
		buf,
		plin.WORD(len(buf)*int(unsafe.Sizeof(buf[0]))),
		&count,
	)
	if err != nil {
		return nil, err
	}

	for i := 0; i < int(count); i++ {

		hw := buf[i]

		// ✅ Correct native sizes (matches PLIN API)
		var hwType plin.BYTE     // <-- FIXED (was WORD)
		var devNo plin.WORD
		var channel plin.BYTE
		var mode plin.WORD

		// --- TYPE ---
		if err := plin.LIN_GetHardwareParam(
			hw,
			plin.TLIN_HARDWAREPARAM_TYPE,
			(*plin.BYTE)(unsafe.Pointer(&hwType)),
			plin.WORD(unsafe.Sizeof(hwType)),
		); err != nil {
			continue
		}

		// --- DEVICE NUMBER ---
		if err := plin.LIN_GetHardwareParam(
			hw,
			plin.TLIN_HARDWAREPARAM_DEVICE_NUMBER,
			(*plin.BYTE)(unsafe.Pointer(&devNo)),
			plin.WORD(unsafe.Sizeof(devNo)),
		); err != nil {
			continue
		}

		// --- CHANNEL ---
		if err := plin.LIN_GetHardwareParam(
			hw,
			plin.TLIN_HARDWAREPARAM_CHANNEL_NUMBER,
			(*plin.BYTE)(unsafe.Pointer(&channel)),
			plin.WORD(unsafe.Sizeof(channel)),
		); err != nil {
			continue
		}

		// --- MODE ---
		if err := plin.LIN_GetHardwareParam(
			hw,
			plin.TLIN_HARDWAREPARAM_MODE,
			(*plin.BYTE)(unsafe.Pointer(&mode)),
			plin.WORD(unsafe.Sizeof(mode)),
		); err != nil {
			continue
		}

		// ✅ switch now compiles (BYTE == BYTE)
		var hwName string
		switch hwType {
		case plin.LIN_HW_TYPE_USB_PRO:
			hwName = "PCAN-USB Pro"

		case plin.LIN_HW_TYPE_USB_PRO_FD:
			hwName = "PCAN-USB Pro FD"

		case plin.LIN_HW_TYPE_PLIN_USB:
			hwName = "PLIN-USB"

		default:
			hwName = "Unknown"
		}

		result = append(result, HardwareInfo{
			Handle:  hw,
			Type:    hwName,
			Device:  int(devNo),
			Channel: int(channel),
		})
	}

	return result, nil
}

func (app *ConsoleApp) readMessage() error {

	var msg plin.TLINRcvMsg

	err := plin.LIN_Read(app.Client, &msg)
	if err != nil {
		return err
	}

	fmt.Printf(
		"ID:%02X Len:%d Data:%v Time:%d\n",
		msg.FrameId,
		msg.Length,
		msg.Data,
		msg.TimeStamp,
	)

	return nil
}

func (app *ConsoleApp) writeMessage(frameID byte, data []byte) error {

	var msg plin.TLINMsg

	msg.FrameId = plin.BYTE(frameID)
	msg.Length = plin.BYTE(len(data))

	for i := range data {
		msg.Data[i] = plin.BYTE(data[i])
	}

	return plin.LIN_Write(
		&app.Client,
		app.Hardware,
		&msg,
	)
}

func (app *ConsoleApp) listHardware() ([]plin.HLINHW, error) {

	buf := make([]plin.HLINHW, 16)
	var count plin.WORD

	err := plin.LIN_GetAvailableHardware(
		buf,
		plin.WORD(len(buf)*int(unsafe.Sizeof(buf[0]))),
		&count,
	)
	if err != nil {
		return nil, err
	}

	return buf[:int(count)], nil
}

func (app *ConsoleApp) menuInput(prompt string) string {
    reader := bufio.NewReader(os.Stdin)

    fmt.Printf("\n * %s", prompt)

    text, _ := reader.ReadString('\n')
    text = strings.TrimSpace(text)
    text = strings.ToUpper(text)

    fmt.Println()
    return text
}

func (app *ConsoleApp) ShowMainMenu() bool {

    fmt.Println("\n\nPLin-API Console:")
    fmt.Println("\t")
    fmt.Println("\t1) View available LIN hardware")
    fmt.Println("\t2) Identify LIN hardware...")
    fmt.Println("\t   ---")

    if app.Hardware == 0 {
        fmt.Println("\t3) Connect to a LIN hardware...")
    } else {
        fmt.Println("\t3) Release LIN hardware")
        fmt.Println("\t   ---")
        fmt.Println("\t4) Global frames table")
        fmt.Println("\t5) Filter status")
        fmt.Println("\t6) Read messages")
        fmt.Println("\t7) Transmit messages")
        fmt.Println("\t8) Status")
        fmt.Println("\t9) Reset")
        fmt.Println("\t10) Versions")
    }

    fmt.Println("\t   ---")
    fmt.Println("\tq) Quit")

    choice := app.menuInput("Select an action: ")

    switch choice {

    case "1":
        app.menuAvailableHw()

    case "2":
        app.menuIdentify()

    case "3":
        if app.Hardware == 0 {
            app.menuConnect()
        } else {
            app.menuDisconnect()
        }

    case "4":
        if app.Hardware != 0 {
            app.menuGlobalFrameTable()
        }

    case "5":
        app.menuFilter()

    case "6":
        app.menuRead()

    case "7":
        app.menuWrite()

    case "8":
        app.menuStatus()

    case "9":
        app.menuReset()

    case "10":
        app.menuVersion()

    case "Q":
        return true

    default:
        app.notify("** Invalid choice **")
    }

    return false
}

func (app *ConsoleApp) menuAvailableHw() {
    app.displayAvailableConnection()
    app.waitEnter()
}

func (app *ConsoleApp) menuIdentify() {

    for {
        app.displayAvailableConnection()

        choice := app.menuInput("Select hardware to identify (Q=exit): ")

        if choice == "Q" {
            return
        }

        idx, err := strconv.Atoi(choice)
        if err != nil {
            app.notify("Invalid selection")
            continue
        }

        err = plin.LIN_IdentifyHardware(plin.HLINHW(idx))
        if err != nil {
            app.notify("Identify failed")
        } else {
            app.notify(fmt.Sprintf("Blinking LED for hardware %d", idx))
        }
    }
}

func (app *ConsoleApp) menuConnect() {

    for {
        app.displayAvailableConnection()

        choice := app.menuInput("Select hardware to connect (Q=exit): ")
        if choice == "Q" {
            return
        }

        hwID, err := strconv.Atoi(choice)
        if err != nil {
            app.notify("Invalid hardware")
            continue
        }

        fmt.Println(" Available hardware mode:")
        fmt.Println("\t1) Master")
        fmt.Println("\t2) Slave")

        modeChoice := app.menuInput("Specify connection mode: ")

        var mode plin.TLINHardwareMode

        switch modeChoice {
        case "1":
            mode = plin.TLIN_HARDWAREMODE_MASTER
        case "2":
            mode = plin.TLIN_HARDWAREMODE_SLAVE
        default:
            app.notify("Invalid mode")
            continue
        }

        baudChoice := app.menuInput("Specify baudrate [19200]: ")

        baud := plin.WORD(19200)
        if baudChoice != "" {
            v, err := strconv.Atoi(baudChoice)
            if err == nil {
                baud = plin.WORD(v)
            }
        }

        err = app.doLinConnect(plin.HLINHW(hwID), mode, baud)
        if err != nil {
            app.notify("Connection failed")
        } else {
            app.notify("Connection successful")
            return
        }
    }
}

func (app *ConsoleApp) menuDisconnect() {

    if err := app.doLinDisconnect(); err != nil {
        app.notify("Disconnection failed")
    } else {
        app.notify("Disconnection successful")
    }
}

func (app *ConsoleApp) menuFilter() {

    //var mask uint64

    err := plin.LIN_GetClientFilter(app.Client, app.Hardware, &app.Mask)
    if err != nil {
        app.notify("Failed to read filter")
        return
    }

    fmt.Printf("Filter mask:\n\t%064b\n", app.Mask)
    app.waitEnter()
}

func (app *ConsoleApp) menuReset() {

    err := plin.LIN_ResetClient(&app.Client)
    if err != nil {
        app.notify("Reset failed")
    } else {
        app.notify("Receive Queue successfully flushed")
    }
}

func (app *ConsoleApp) menuVersion() {

    var ver plin.TLINVersion

    err := plin.LIN_GetVersion(&ver)
    if err != nil {
        app.notify("Version read failed")
        return
    }

    fmt.Printf(
        "API Version: %d.%d.%d.%d\n",
        ver.Major,
        ver.Minor,
        ver.Build,
        ver.Revision,
    )

    app.waitEnter()
}

func (app *ConsoleApp) menuGlobalFrameTable() {
	bQuit := false

	for !bQuit {
		// Clear screen and display header + global frames
		app.clearMenu()
		app.displayMenuHeader()
		app.displayGlobalFramesTable(true)

		// Display menu options
		fmt.Println("\t")
		fmt.Println("\t1) Configure default example configuration for global frames table")
		app.displayMenuExit("")

		choice := app.displayMenuInput("Select an action (Q=quit): ")
		switch choice {
		case "Q":
			bQuit = true

		case "1":
			// Retrieve global frames table
			frames, _ := app.readFrameTableFromHw()

			// Disable all frames and set default CST & length
			for i := range frames {
				frame := &frames[i] // pointer to modify in place
				frame.ChecksumType = plin.TLIN_CHECKSUMTYPE_ENHANCED
				frame.Direction = plin.TLIN_DIRECTION_DISABLED

				// Set lengths according to LIN 1.2 specification
				switch {
				case frame.FrameId <= 0x1F:
					frame.Length = 2
				case frame.FrameId <= 0x2F:
					frame.Length = 4
				case frame.FrameId <= 0x3F:
					frame.Length = 8
				}

				// Apply frame entry to hardware
				if err := plin.LIN_SetFrameEntry(&app.Client, app.Hardware, frame); err != nil {
					app.notify(fmt.Sprintf("Failed to set frame 0x%X: %v", frame.FrameId, err))
				}
			}

			// Reset filter mask (all frames disabled)
			app.Mask = 0
			if err := plin.LIN_SetClientFilter(app.Client, app.Hardware, &app.Mask); err != nil {
				app.notify(fmt.Sprintf("Failed to update client filter mask: %v", err))
			}

			// Determine publisher/subscriber directions depending on hardware mode
			var directionPub, directionSub plin.TLINDirection
			if app.HwMode == plin.TLIN_HARDWAREMODE_MASTER {
				directionPub = plin.TLIN_DIRECTION_PUBLISHER
				directionSub = plin.TLIN_DIRECTION_SUBSCRIBER
			} else {
				directionPub = plin.TLIN_DIRECTION_SUBSCRIBER
				directionSub = plin.TLIN_DIRECTION_PUBLISHER
			}

			// Apply example frame entries
			success := true
			success = success && app.setFrameEntry(0x01, directionSub, plin.TLIN_CHECKSUMTYPE_ENHANCED, 8)
			success = success && app.setFrameEntry(0x02, directionSub, plin.TLIN_CHECKSUMTYPE_ENHANCED, 2)
			success = success && app.setFrameEntry(0x03, directionSub, plin.TLIN_CHECKSUMTYPE_ENHANCED, 8)
			success = success && app.setFrameEntry(0x05, directionPub, plin.TLIN_CHECKSUMTYPE_ENHANCED, 2)
			success = success && app.setFrameEntry(0x3C, directionPub, plin.TLIN_CHECKSUMTYPE_CLASSIC, 8)
			success = success && app.setFrameEntry(0x3D, directionSub, plin.TLIN_CHECKSUMTYPE_CLASSIC, 8)

			if success {
				app.notify("Global frames table successfully configured")
			} else {
				app.notify("Global frames table configuration failed")
			}

		default:
			app.notify("** Invalid choice **")
		}
	}
}

func (app *ConsoleApp) menuRead() {
	bQuit := false
	var listMsg []string // buffer of received messages

	for !bQuit {
		app.clearMenu()
		app.displayMenuHeader()
		fmt.Println("\t0-9) Read a specified number of messages")
		app.displayMenuExit("")

		if len(listMsg) > 0 {
			fmt.Println("\n * Received messages:")
			fmt.Println("   ID\tLength\tData\tTimestamp\tDirection\tErrors")
			fmt.Println("   -----------------------------------------------------------")
			for _, msg := range listMsg {
				fmt.Println("\n - " + msg)
			}
			listMsg = nil
		}

		choice := app.displayMenuInput("Number of messages to read [1]: ")
		if choice == "Q" {
			bQuit = true
			continue
		}

		numRead := 1
		if choice != "" && choice != "1" {
			v, err := strconv.Atoi(choice)
			if err == nil && v > 0 {
				numRead = v
			} else {
				app.notify("Invalid input, reading 1 message")
			}
		}

		if numRead == 1 {
			var msg plin.TLINRcvMsg
			err := plin.LIN_Read(app.Client, &msg)
			if err == nil {
				listMsg = append(listMsg, app.getFormattedRcvMsg(msg))
			} else {
				app.displayNotification(fmt.Sprintf("Read error: %v", err), 0.5)
			}
		} else {
			msgArray := make([]plin.TLINRcvMsg, numRead)
			msgCount := 0

			for i := 0; i < numRead; i++ {
				err := plin.LIN_Read(app.Client, &msgArray[i])
				if err != nil {
					break
				}
				msgCount++
			}

			if msgCount == 0 {
				app.displayNotification("No messages available", 0.5)
			} else {
				for i := 0; i < msgCount; i++ {
					listMsg = append(listMsg, app.getFormattedRcvMsg(msgArray[i]))
				}
			}
		}
	}
}

func (app *ConsoleApp) menuWrite() {
	bQuit := false
	bShowFrames := true

	for !bQuit {
		app.clearMenu()
		app.displayMenuHeader()
		fmt.Println("\t0x00-0x3F) Write a LIN message with a valid frame ID")
		fmt.Println("\tt) Toggle display of global frames table")
		app.displayMenuExit("")

		if bShowFrames {
			app.displayGlobalFramesTable(false)
		}

		choice := app.displayMenuInput("Action or Frame ID (hex): ")

		if choice == "Q" {
			bQuit = true
			continue
		} else if choice == "T" {
			bShowFrames = !bShowFrames
			continue
		}

		// Parse Frame ID
		frameID64, err := strconv.ParseUint(choice, 16, 8)
		if err != nil {
			app.notify("Invalid Frame ID")
			continue
		}
		frameID := plin.BYTE(frameID64)

		// Get frame entry from hardware
		var frame plin.TLINFrameEntry
		frame.FrameId = frameID
		err = plin.LIN_GetFrameEntry(app.Hardware, &frame)
		if err != nil {
			app.notify(fmt.Sprintf("Failed to get frame entry: %v", err))
			continue
		}

		// Prepare message
		var msg plin.TLINMsg
		msg.FrameId = frame.FrameId
		msg.Length = frame.Length
		msg.Direction = frame.Direction
		msg.ChecksumType = frame.ChecksumType

		// Fill data if Publisher
		if msg.Direction == plin.TLIN_DIRECTION_PUBLISHER {
			for i := plin.BYTE(0); i < frame.Length; i++ {
				dataStr := app.displayMenuInput(fmt.Sprintf("Data[%d] (hex): ", i+1))
				v, parseErr := strconv.ParseUint(dataStr, 16, 8)
				if parseErr != nil {
					app.notify("Invalid input, using 0")
					v = 0
				}
				msg.Data[i] = plin.BYTE(v)
			}
		}

		// Write message or update byte array depending on master/slave
		if app.HwMode == plin.TLIN_HARDWAREMODE_MASTER {
			// Set protected ID and calculate checksum
			nPID := frameID
			err = plin.LIN_GetPID(&nPID)
			if err != nil {
				app.notify("Failed to get PID")
			}
			msg.FrameId = nPID

			err = plin.LIN_CalculateChecksum(&msg)
			if err != nil {
				app.notify("Failed to calculate checksum")
			}

			err = plin.LIN_Write(&app.Client, app.Hardware, &msg)
		} else {
			// Slave: update the byte array directly
			err = plin.LIN_UpdateByteArray(&app.Client, app.Hardware, frameID, 0, msg.Length, &msg.Data[0])
		}

		if err == nil {
			app.displayNotification("Message successfully written", 1.0)
		} else {
			app.notify(fmt.Sprintf("Failed to write message: %v", err))
		}
	}
}

func (app *ConsoleApp) menuStatus() {
	var status plin.TLINHardwareStatus
	err := plin.LIN_GetStatus(app.Hardware, &status)

	if err != nil {
		app.notify(fmt.Sprintf("Failed to read status: %v", err))
		return
	}

	switch status.Status {
	case plin.TLIN_HARDWARESTATE_ACTIVE:
		app.displayNotification("Bus: Active", 1.0)
	case plin.TLIN_HARDWARESTATE_AUTOBAUDRATE:
		app.displayNotification("Hardware: Baudrate Detection", 1.0)
	case plin.TLIN_HARDWARESTATE_NOT_INITIALIZED:
		app.displayNotification("Hardware: Not Initialized", 1.0)
	case plin.TLIN_HARDWARESTATE_SHORT_GROUND:
		app.displayNotification("Bus-Line: Shorted Ground", 1.0)
	case plin.TLIN_HARDWARESTATE_SLEEP:
		app.displayNotification("Bus: Sleep", 1.0)
	default:
		app.displayNotification("Bus: Unknown Status", 1.0)
	}
}

func (app *ConsoleApp) notify(msg string) {
    fmt.Println(msg)
}


func (app *ConsoleApp) waitEnter() {
    fmt.Println("Press ENTER to continue...")
    bufio.NewReader(os.Stdin).ReadBytes('\n')
}

func getFrameDirectionAsString(direction plin.TLINDirection) string {
    switch direction {
    case plin.TLIN_DIRECTION_DISABLED:
        return "Disabled"
    case plin.TLIN_DIRECTION_PUBLISHER:
        return "Publisher"
    case plin.TLIN_DIRECTION_SUBSCRIBER:
        return "Subscriber"
    case plin.TLIN_DIRECTION_SUBSCRIBER_AUTOLENGTH:
        return "Subscriber Automatic Length"
    default:
        return fmt.Sprintf("Unknown (%d)", direction)
    }
}

func getFrameCSTAsString(checksumType plin.TLINChecksumType) string {
	switch checksumType {
	case plin.TLIN_CHECKSUMTYPE_AUTO:
		return "Auto"
	case plin.TLIN_CHECKSUMTYPE_CLASSIC:
		return "Classic"
	case plin.TLIN_CHECKSUMTYPE_CUSTOM:
		return "Custom"
	case plin.TLIN_CHECKSUMTYPE_ENHANCED:
		return "Enhanced"
	default:
		return fmt.Sprintf("Unknown (%d)", checksumType)
	}
}

func (app *ConsoleApp) getFormattedRcvMsg(msg plin.TLINRcvMsg) string {
	// Handle non-standard message types
	switch msg.Type {
	case plin.TLIN_MSGTYPE_STANDARD:
		// standard frame, do nothing
	default:
		switch msg.Type {
		case plin.TLIN_MSGTYPE_BUS_SLEEP:
			return "Bus Sleep status message"
		case plin.TLIN_MSGTYPE_BUS_WAKEUP:
			return "Bus WakeUp status message"
		case plin.TLIN_MSGTYPE_AUTOBAUDRATE_TIMEOUT:
			return "Auto-baudrate Timeout status message"
		case plin.TLIN_MSGTYPE_AUTOBAUDRATE_REPLY:
			return "Auto-baudrate Reply status message"
		case plin.TLIN_MSGTYPE_OVERRUN:
			return "Bus Overrun status message"
		case plin.TLIN_MSGTYPE_QUEUE_OVERRUN:
			return "Queue Overrun status message"
		default:
			return "Non standard message"
		}
	}

	// Format Data field
	dataStr := ""
	for i := 0; i < int(msg.Length); i++ {
		dataStr += fmt.Sprintf("%#x ", msg.Data[i])
	}
	dataStr = strings.TrimSpace(dataStr)

	// Format Error field
	errorStr := ""
	flags := msg.ErrorFlags

	if flags&plin.TLINMsgErrors(plin.TLIN_MSGERROR_CHECKSUM) != 0 {
		errorStr += "Checksum,"
	}
	if flags&plin.TLINMsgErrors(plin.TLIN_MSGERROR_GROUND_SHORT) != 0 {
		errorStr += "GroundShort,"
	}
	if flags&plin.TLINMsgErrors(plin.TLIN_MSGERROR_ID_PARITY_BIT_0) != 0 {
		errorStr += "IdParityBit0,"
	}
	if flags&plin.TLINMsgErrors(plin.TLIN_MSGERROR_ID_PARITY_BIT_1) != 0 {
		errorStr += "IdParityBit1,"
	}
	if flags&plin.TLINMsgErrors(plin.TLIN_MSGERROR_INCONSISTENT_SYNCH) != 0 {
		errorStr += "InconsistentSynch,"
	}
	if flags&plin.TLINMsgErrors(plin.TLIN_MSGERROR_OTHER_RESPONSE) != 0 {
		errorStr += "OtherResponse,"
	}
	if flags&plin.TLINMsgErrors(plin.TLIN_MSGERROR_SLAVE_NOT_RESPONDING) != 0 {
		errorStr += "SlaveNotResponding,"
	}
	if flags&plin.TLINMsgErrors(plin.TLIN_MSGERROR_SLOT_DELAY) != 0 {
		errorStr += "SlotDelay,"
	}
	if flags&plin.TLINMsgErrors(plin.TLIN_MSGERROR_TIMEOUT) != 0 {
		errorStr += "Timeout,"
	}
	if flags&plin.TLINMsgErrors(plin.TLIN_MSGERROR_VBAT_SHORT) != 0 {
		errorStr += "VBatShort,"
	}

	if flags == 0 {
		errorStr = "O.k."
	} else {
		errorStr = strings.TrimRight(errorStr, ",")
	}

	pid := app.PIDs[msg.FrameId]

	return fmt.Sprintf("%#x\t%d\t%s\t...%d\t%s\t%s",
		pid,
		msg.Length,
		dataStr,
		msg.TimeStamp,
		getFrameDirectionAsString(msg.Direction),
		errorStr,
	)
}

func (app *ConsoleApp) setFrameEntry(frameID byte, direction plin.TLINDirection, checksumType plin.TLINChecksumType, length byte) bool {
	frame := plin.TLINFrameEntry{
		FrameId:      plin.BYTE(frameID),      // OK
		Direction:    direction,
		ChecksumType: checksumType,
		Length:       plin.BYTE(length),       // ✅ cast explicitly
		Flags:        plin.FRAME_FLAG_RESPONSE_ENABLE,
	}

	if err := plin.LIN_SetFrameEntry(&app.Client, app.Hardware, &frame); err != nil {
		app.notify(fmt.Sprintf("Failed to set frame: %v", err))
		return false
	}

	// update filter mask
	app.Mask |= 1 << frameID
	if err := plin.LIN_SetClientFilter(app.Client, app.Hardware, &app.Mask); err != nil {
		app.notify(fmt.Sprintf("Failed to update filter mask: %v", err))
		return false
	}

	return true
}

func (app *ConsoleApp) displayAvailableConnection() {
	hwList, err := app.getAvailableHardware()
	if err != nil {
		fmt.Println("\t<Error retrieving hardware>")
		return
	}

	fmt.Println("List of available LIN hardware:")
	if len(hwList) == 0 {
		fmt.Println("\t<No hardware found>")
	} else {
		for idx, hw := range hwList {
			isConnected := ""
			if app.Hardware == hw.Handle {
				modeStr := "slave"
				if app.HwMode == plin.TLIN_HARDWAREMODE_MASTER {
					modeStr = "master"
				}
				isConnected = fmt.Sprintf("(connected as %s, %d)", modeStr, app.HwBaudrate)
			}
			fmt.Printf("\t%d) %s - dev. %d, chan. %d %s\n", idx, hw.Type, hw.Device, hw.Channel, isConnected)
		}
	}
}

func (app *ConsoleApp) displayGlobalFramesTable(showDisabled bool) {
	frames, _ := app.readFrameTableFromHw()

	fmt.Println("\n* Global Frames Table:\n")
	fmt.Println("ID\tPID\tDirection\t\tLength\tCST")
	fmt.Println("------------------------------------------------------------")

	for _, frame := range frames {
		if frame.Direction != plin.TLIN_DIRECTION_DISABLED || showDisabled {
			pid := app.PIDs[frame.FrameId]
			fmt.Printf("%#x\t%#x\t%s\t%d\t%s\n",
				frame.FrameId,
				pid,
				getFrameDirectionAsString(frame.Direction),
				frame.Length,
				getFrameCSTAsString(frame.ChecksumType),
			)
		}
	}
}

func (app *ConsoleApp) displayMenuHeader() {
	fmt.Println("\n\nPLin-API Console:")
}

func (app *ConsoleApp) displayMenuExit(text string) {
	if text == "" {
		text = "q) Exit menu"
	}
	fmt.Printf("\t%s\n", text)
}

func (app *ConsoleApp) displayMenuInput(prompt string) string {
	fmt.Printf("\n * %s", prompt)
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)
	return strings.ToUpper(text)
}

func (app *ConsoleApp) displayNotification(text string, waitSeconds float64) {
	fmt.Printf("\t%s\n", text)
	time.Sleep(time.Duration(waitSeconds * float64(time.Second)))
}

func (app *ConsoleApp) clearMenu() {
	cmd := ""
	args := []string{}

	if runtime.GOOS == "windows" {
		cmd = "cmd"
		args = []string{"/c", "cls"}
	} else {
		cmd = "clear"
	}

	if cmd != "" {
		execCmd := exec.Command(cmd, args...)
		execCmd.Stdout = os.Stdout
		_ = execCmd.Run()
	}
}


func main() {

    app, err := NewConsoleApp()
	if err != nil {
		panic(err)
	}

	defer app.uninitialize()

	hw, err := app.listHardware()
	if err != nil {
		panic(err)
	}

	fmt.Println("Found hardware:", hw)

    for {
        if app.ShowMainMenu() {
            break
        }
    }
}
