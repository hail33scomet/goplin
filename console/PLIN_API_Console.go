package main

import (
	"fmt"
	"unsafe"
	"github.com/hail33scomet/goplin"
)

type ConsoleApp struct {
	Client     plin.HLINCLIENT
	Hardware   plin.HLINHW
	HwMode     plin.TLINHardwareMode
	HwBaudrate plin.WORD
	Mask       plin.UINT64
	PIDs       map[plin.BYTE]int
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

func (PLINApp *ConsoleApp) initialize() error {

	PLINApp.Client = 0
	PLINApp.Hardware = 0
	PLINApp.HwMode = plin.TLIN_HARDWAREMODE_NONE
	PLINApp.HwBaudrate = 0
	PLINApp.Mask = 0xFFFFFFFFFFFFFFFF

	// Initialize PID mapping
	for i := 0; i < 64; i++ {
		//frame := plin.BYTE(i)
		var pid plin.BYTE

		err := plin.LIN_GetPID(&pid)
		if err != nil {
			return err
		}
		PLINApp.PIDs[pid] = i
	}

	return nil
}

func (PLINApp *ConsoleApp) uninitialize() error {

	if PLINApp.Client != 0 {

		err := PLINApp.doLinDisconnect()
		if err != nil {
			return err
		}

		err = plin.LIN_RemoveClient(&PLINApp.Client)
		if err != nil {
			return err
		}

		PLINApp.Client = 0
	}

	return nil
}

func (PLINApp *ConsoleApp) doLinConnect(
	hw plin.HLINHW,
	mode plin.TLINHardwareMode,
	baud plin.WORD,
) error {

	if PLINApp.Hardware != 0 {
		if err := PLINApp.doLinDisconnect(); err != nil {
			return err
		}
	}

	if PLINApp.Client == 0 {

		err := plin.LIN_RegisterClient(
			"PLIN-API Console",
			0,
			&PLINApp.Client,
		)

		if err != nil {
			return err
		}
	}

	err := plin.LIN_ConnectClient(&PLINApp.Client, hw)
	if err != nil {
		return err
	}

	PLINApp.Hardware = hw

	
		err = plin.LIN_InitializeHardware(
		&PLINApp.Client,
		PLINApp.Hardware,
		plin.WORD(mode),
		baud,
	)

	if err != nil {
		return err
	}

	PLINApp.HwMode = mode
	PLINApp.HwBaudrate = baud

	return plin.LIN_SetClientFilter(
		PLINApp.Client,
		PLINApp.Hardware,
		PLINApp.Mask,
	)
}

func (PLINApp *ConsoleApp) doLinDisconnect() error {

	if PLINApp.Hardware == 0 {
		return nil
	}

	err := plin.LIN_DisconnectClient(PLINApp.Client, PLINApp.Hardware)
	if err != nil {
		return err
	}

	PLINApp.Hardware = 0

	return nil
}

func (PLINApp *ConsoleApp) readMessage() error {

	var msg plin.TLINRcvMsg

	err := plin.LIN_Read(PLINApp.Client, &msg)
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

func (PLINApp *ConsoleApp) writeMessage(frameID byte, data []byte) error {

	var msg plin.TLINMsg

	//msg.FrameId = frameID
	msg.FrameId = plin.BYTE(frameID)
	msg.Length = plin.BYTE(len(data))

	for i := range data {
		//msg.Data[i] = data[i]
		msg.Data[i] = plin.BYTE(data[i])
	}

	return plin.LIN_Write(
		&PLINApp.Client,
		PLINApp.Hardware,
		&msg,
	)
}

func (PLINApp *ConsoleApp) listHardware() ([]plin.HLINHW, error) {

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


func main() {

	fmt.Println("PLIN Go Console Example")

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

}