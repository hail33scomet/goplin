package plin

import (
	"sync"
	"syscall"
	"unsafe"
	"bytes"
	"fmt"
)

var (
	funcMap = map[string]**syscall.Proc{
		"LIN_GetSystemTime":  &procLINGetSystemTime,
		"LIN_GetResponseRemap": &procLINGetResponseRemap,
		"LIN_SetResponseRemap": &procLINSetResponseRemap,
		"LIN_GetTargetTime": &procLINGetTargetTime,
		"LIN_GetPID": &procLINGetPID,
		"LIN_GetErrorText": &procLINGetErrorText, 
		"LIN_GetVersionInfo": &procLINGetVersionInfo,
		"LIN_GetVersion": &procLINGetVersion,
		"LIN_CalculateChecksum": &procLINCalculateChecksum,
		"LIN_GetStatus": &procLINGetStatus,
		"LIN_StartAutoBaud": &procLINStartAutoBaud,
		"LIN_XmtDynamicWakeUp": &procLINXmtDynamicWakeUp,
		"LIN_XmtWakeUp": &procLINXmtWakeUp,
		"LIN_ResumeSchedule": &procLINResumeSchedule,
		"LIN_SuspendSchedule": &procLINSuspendSchedule,
		"LIN_StartSchedule": &procLINStartSchedule,
		"LIN_SetScheduleBreakPoint": &procLINSetScheduleBreakPoint,
		"LIN_DeleteSchedule": &procLINDeleteSchedule,
		"LIN_GetSchedule": &procLINGetSchedule,
		"LIN_SetSchedule": &procLINSetSchedule,
		"LIN_ResumeKeepAlive": &procLINResumeKeepAlive,
		"LIN_SuspendKeepAlive": &procLINSuspendKeepAlive,
		"LIN_StartKeepAlive": &procLINStartKeepAlive,
		"LIN_UpdateByteArray": &procLINUpdateByteArray,
		"LIN_GetFrameEntry": &procLINGetFrameEntry,
		"LIN_SetFrameEntry": &procLINSetFrameEntry,
		"LIN_RegisterFrameId": &procLINRegisterFrameId,
		"LIN_IdentifyHardware": &procLINIdentifyHardware,
		"LIN_ResetHardwareConfig": &procLINResetHardwareConfig,
		"LIN_ResetHardware": &procLINResetHardware,
		"LIN_GetHardwareParam": &procLINGetHardwareParam,
		"LIN_SetHardwareParam": &procLINSetHardwareParam,
		"LIN_GetAvailableHardware": &procLINGetAvailableHardware,
		"LIN_InitializeHardware": &procLINInitializeHardware,
		"LIN_Write": &procLINWrite,
		"LIN_ReadMulti": &procLINReadMulti,
		"LIN_Read": &procLINRead,
		"LIN_GetClientFilter": &procLINGetClientFilter,
		"LIN_SetClientFilter": &procLINSetClientFilter,
		"LIN_GetClientParam": &procLINGetClientParam,
		"LIN_SetClientParam": &procLINSetClientParam,
		"LIN_ResetClient": &procLINResetClient,
		"LIN_DisconnectClient": &procLINDisconnectClient,
		"LIN_ConnectClient": &procLINConnectClient,
		"LIN_RemoveClient": &procLINRemoveClient,
		"LIN_RegisterClient": &procLINRegisterClient,
	}

	plin                  *syscall.DLL
	procLINGetSystemTime  *syscall.Proc
	procLINGetResponseRemap *syscall.Proc
	procLINSetResponseRemap *syscall.Proc
	procLINGetTargetTime *syscall.Proc
	procLINGetPID *syscall.Proc
	procLINGetErrorText *syscall.Proc
	procLINGetVersionInfo *syscall.Proc
	procLINGetVersion *syscall.Proc
	procLINCalculateChecksum *syscall.Proc
	procLINGetStatus *syscall.Proc
	procLINStartAutoBaud *syscall.Proc
	procLINXmtDynamicWakeUp *syscall.Proc
	procLINXmtWakeUp *syscall.Proc
	procLINResumeSchedule *syscall.Proc
	procLINSuspendSchedule *syscall.Proc
	procLINStartSchedule *syscall.Proc
	procLINSetScheduleBreakPoint *syscall.Proc
	procLINDeleteSchedule *syscall.Proc
	procLINGetSchedule *syscall.Proc
	procLINSetSchedule *syscall.Proc
	procLINResumeKeepAlive *syscall.Proc
	procLINSuspendKeepAlive *syscall.Proc
	procLINStartKeepAlive *syscall.Proc
	procLINUpdateByteArray *syscall.Proc
	procLINGetFrameEntry *syscall.Proc
	procLINSetFrameEntry *syscall.Proc
	procLINRegisterFrameId *syscall.Proc
	procLINIdentifyHardware *syscall.Proc
	procLINResetHardwareConfig *syscall.Proc
	procLINResetHardware *syscall.Proc
	procLINGetHardwareParam *syscall.Proc
	procLINSetHardwareParam *syscall.Proc
	procLINGetAvailableHardware *syscall.Proc
	procLINInitializeHardware *syscall.Proc
	procLINWrite *syscall.Proc
	procLINReadMulti *syscall.Proc
	procLINRead *syscall.Proc
	procLINGetClientFilter *syscall.Proc
	procLINSetClientFilter *syscall.Proc
	procLINGetClientParam *syscall.Proc
	procLINSetClientParam *syscall.Proc
	procLINResetClient *syscall.Proc
 	procLINDisconnectClient *syscall.Proc
	procLINConnectClient *syscall.Proc
	procLINRemoveClient *syscall.Proc
	procLINRegisterClient *syscall.Proc
)

var (
	initOnce sync.Once
	InitErr  error
	plinDLL  *syscall.DLL
)

func initDLL() error {
	initOnce.Do(func() {

		plinDLL, InitErr = syscall.LoadDLL("PLinApi.dll")
		if InitErr != nil {
			InitErr = fmt.Errorf("failed to load PLinApi.dll: %w", InitErr)
			return
		}

		for name, proc := range funcMap {

			p, err := findProcWithStdcall(plinDLL, name)
			if err != nil {
				InitErr = err
				return
			}

			*proc = p
		}
	})

	return InitErr
}

func findProcWithStdcall(dll *syscall.DLL, name string) (*syscall.Proc, error) {

	// Try undecorated first
	if p, err := dll.FindProc(name); err == nil {
		return p, nil
	}

	// Try stdcall decorations (common stack sizes)
	for _, suffix := range []int{
		4, 8, 12, 16, 20, 24, 28, 32,
		36, 40, 44, 48, 52, 56, 60, 64,
	} {
		decorated := fmt.Sprintf("%s@%d", name, suffix)

		if p, err := dll.FindProc(decorated); err == nil {
			return p, nil
		}
	}

	return nil, fmt.Errorf("PLIN procedure not found: %s", name)
}

func ensureInit() {
	if err := initDLL(); err != nil {
		panic(err)
	}
}

type PLINError struct {
	Code TLINError
}

func (e PLINError) Error() string {
	const errBufSize = 256
	buf := make([]byte, errBufSize)

	LIN_GetErrorText(
		e.Code,
		0,
		&buf[0], // already *byte
		WORD(len(buf)),
	)

	n := bytes.IndexByte(buf, 0)
	if n == -1 {
		n = len(buf)
	}

	return string(buf[:n])
}

func checkErr(r1, _ uintptr, _ error) error {
	if r1 != uintptr(TLIN_ERROR_OK) {
		return PLINError{Code: TLINError(r1)}
	}
	return nil
}

/*
	Registers a Client at the LIN Manager. Creates a Client handle and 
	allocates the Receive Queue (only one per Client). 
	
	Remarks: 
		The hWnd parameter can be zero for DOS Box Clients. 
		The Client does not receive any messages until LIN_RegisterFrameId() 
		or LIN_SetClientFilter() is called.
	
	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED, 
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS
	
	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE,
		TLIN_ERROR_OUT_OF_RESOURCE
	
	Parameters: 
		strName     : Name of the Client (python string) 
		hWnd        : Window handle of the Client (only for information purposes) (c_ulong) 
		hClient     : Client handle buffer (HLINCLIENT) 
	
	Returns:
            A LIN Error Code
*/

func LIN_RegisterClient(strName string, hWnd HLINHW, hClient *HLINCLIENT) error {
	//ensureInit()
	namePtr, err := syscall.BytePtrFromString(strName)
	if err != nil {
        return err
    }

	return checkErr(procLINRegisterClient.Call(
		uintptr(unsafe.Pointer(namePtr)), 
		uintptr(hWnd), 
		uintptr(unsafe.Pointer(hClient))))
}

/*
	Removes a Client from the Client list of the LIN Manager. 
	Frees all resources (receive queues, message counters, etc.). 

	Remarks:
		If the Client was a Boss-Client for one or more Hardware, 
		the Boss-Client property for those Hardware will be set to 
		INVALID_LIN_HANDLE.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED, 
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_CLIENT 

	Parameters:
		hClient     : Handle of the Client  (HLINCLIENT) 

	Returns:
		A LIN Error Code
*/

func LIN_RemoveClient(hClient *HLINCLIENT) error {
	return checkErr(procLINRemoveClient.Call(
		uintptr(unsafe.Pointer(hClient))))
} 

/*
	Connects a Client to a Hardware.
	The Hardware is assigned by its Handle.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_CLIENT, 
		TLIN_ERROR_ILLEGAL_HARDWARE

	Parameters:
		hClient     : Handle of the Client  (HLINCLIENT) 
		hHw         : Handle of the Hardware to be connected (HLINHW)

	Returns:
		A LIN Error Code
*/

func LIN_ConnectClient(hClient *HLINCLIENT, hHw HLINHW) error {
	return checkErr(procLINConnectClient.Call(
		uintptr(unsafe.Pointer(hClient)), 
		uintptr(hHw)))
} 

/*
	Disconnects a Client from a Hardware.

	Remarks:
		No more messages will be received by this Client from this Hardware.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_CLIENT, 
		TLIN_ERROR_ILLEGAL_HARDWARE

	Parameters:
		hClient     : Handle of the Client  (HLINCLIENT) 
		hHw         : Handle of the Hardware to be disconnected (HLINHW)

	Returns:
		A LIN Error Code
*/
func LIN_DisconnectClient(hclient HLINCLIENT, hw HLINHW) error {
	return checkErr(procLINDisconnectClient.Call(
		uintptr(hclient),
		uintptr(hw),
	))
}

/*
	Flushes the Receive Queue of the Client and resets its counters.

	Remarks:
		No more messages will be received by this Client from this Hardware.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_CLIENT

	Parameters:
		hClient     : Handle of the Client  (HLINCLIENT) 

	Returns:
		A LIN Error Code
*/

func LIN_ResetClient(hClient *HLINCLIENT) error {
	return checkErr(procLINResetClient.Call(
		uintptr(unsafe.Pointer(hClient))))
} 

/*
	Sets a Client parameter to a given value.

	Remarks:
		Allowed TLINClientParam                 Parameter
		values in wParam:                       type:       Description:
		------------------------               ----------  ------------------------------
		TLIN_CLIENTPARAM_RECEIVE_STATUS_FRAME   c_int       0 = Status Frames deactivated,
															otherwise active

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_TYPE, TLIN_ERROR_WRONG_PARAM_VALUE, 
		TLIN_ERROR_ILLEGAL_CLIENT

	Parameters:
		hClient     : Handle of the Client  (HLINCLIENT) 
		wParam      : Parameter to set (TLINClientParam)
		dwValue     : Parameter value (see remarks)

	Returns:
		A LIN Error Code
*/

func LIN_SetClientParam(hClient *HLINCLIENT, wParam TLINClientParam, dwValue BYTE) error {
	return checkErr(procLINSetClientParam.Call(
		uintptr(unsafe.Pointer(hClient)), 
		uintptr(wParam), 
		uintptr(dwValue)))
} 

/*
	Gets a Client parameter.

	Remarks:
		Allowed TLINClientParam                 Parameter
		values in wParam:                       type:       Description:
		-------------------------              ----------  ------------------------------
		TLIN_CLIENTPARAM_NAME                   string      Name of the Client
		TLIN_CLIENTPARAM_MESSAGE_ON_QUEUE       c_int       Unread messages in the Receive Queue
		TLIN_CLIENTPARAM_WINDOW_HANDLE          c_int       Window handle of the Client application 
															(can be zero for DOS Box Clients)
		TLIN_CLIENTPARAM_CONNECTED_HARDWARE     HLINHW[]    Array of Hardware Handles connected by a Client 
															The first item in the array refers to the 
															amount of handles. So [*] = Total handles + 1
		TLIN_CLIENTPARAM_TRANSMITTED_MESSAGES   c_int       Number of transmitted messages
		TLIN_CLIENTPARAM_RECEIVED_MESSAGES      c_int       Number of received messages
		TLIN_CLIENTPARAM_RECEIVE_STATUS_FRAME   c_int       0 = Status Frames deactivated, otherwise active

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_TYPE, TLIN_ERROR_WRONG_PARAM_VALUE, 
		TLIN_ERROR_ILLEGAL_CLIENT, TLIN_ERROR_BUFFER_INSUFFICIENT

	Parameters:
		hClient     : Handle of the Client  (HLINCLIENT) 
		wParam      : Parameter to get (TLINClientParam)
		pBuff       : Buffer for the parameter value (see remarks)
		wBuffSize   : Buffer size in bytes (c_ushort)

	Returns:
		A LIN Error Code
*/
			
func LIN_GetClientParam(hClient *HLINCLIENT, wParam TLINClientParam, pBuff *BYTE, wBuffSize WORD) error {
	return checkErr(procLINGetClientParam.Call(
		uintptr(unsafe.Pointer(hClient)), 
		uintptr(wParam), 
		uintptr(unsafe.Pointer(pBuff)), 
		uintptr(wBuffSize)))
} 

/*
	Sets the filter of a Client and modifies the filter of the connected Hardware.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_CLIENT,
		TLIN_ERROR_ILLEGAL_HARDWARE

	Parameters:
		hClient     : Handle of the Client  (HLINCLIENT) 
		hHw         : Handle of the Hardware (HLINHW)
		iRcvMask    : Filter. Each bit corresponds to a Frame ID (0..63) (c_uint64)

	Returns:
		A LIN Error Code
*/

     
func LIN_SetClientFilter(hclient HLINCLIENT, hw HLINHW, iRcvMask UINT64) error {
	return checkErr(procLINSetClientFilter.Call(
		uintptr(hclient),
		uintptr(hw),
		uintptr(iRcvMask),
	))
}

/*
	Reads the next message/status information from a Client's Receive Queue.

	Remarks:
			The message will be written to 'pMsg'.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_CLIENT, 
		TLIN_ERROR_RCVQUEUE_EMPTY

	Parameters:
		hClient     : Handle of the Client  (HLINCLIENT) 
		pMsg        : Message read buffer (TLINRcvMsg)

	Returns:
		A LIN Error Code
*/  

func LIN_Read(hclient HLINCLIENT, msg *TLINRcvMsg) error {
	return checkErr(procLINRead.Call(
		uintptr(hclient),
		uintptr(unsafe.Pointer(msg)),
	))
}
/*
	Reads several received messages.

	Remarks:
		pMsgBuff must be an array of 'iMaxCount' entries (must have at least 
		a size of iMaxCount * sizeof(TLINRcvMsg) bytes).
		The size 'iMaxCount' of the array = max. messages that can be received.
		The real number of read messages will be returned in 'pCount'.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_CLIENT,
		TLIN_ERROR_RCVQUEUE_EMPTY

	Parameters:
		hClient     : Handle of the Client  (HLINCLIENT) 
		pMsgBuff    : Messages buffer (TLINRcvMsg[])
		iMaxCount   : Maximum number of messages to read (pMsgBuff length) (c_int)
		pCount      : Buffer for the real number of messages read (c_int)

	Returns:
		A LIN Error Code
*/

func LIN_ReadMulti(hClient *HLINCLIENT, pMsgBuff []TLINRcvMsg, iMaxCount SDWORD, pCount *SDWORD) error {
	return checkErr(procLINReadMulti.Call(
		uintptr(unsafe.Pointer(hClient)),  
		uintptr(unsafe.Pointer(&pMsgBuff[0])), 
		uintptr(iMaxCount), 
		uintptr(unsafe.Pointer(pCount))))
}

/*
	The Client 'hClient' transmits a message 'pMsg' to the Hardware 'hHw'.

	Remarks:
			The message is written into the Transmit Queue of the Hardware.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_CLIENT,
		TLIN_ERROR_ILLEGAL_HARDWARE, TLIN_ERROR_ILLEGAL_DIRECTION, 
		TLIN_ERROR_ILLEGAL_LENGTH

	Parameters:
		hClient     : Handle of the Client  (HLINCLIENT) 
		hHw         : Handle of the Hardware (HLINHW)
		pMsg        : Message write buffer (TLINMsg)

	Returns:
		A LIN Error Code
*/
   
func LIN_Write(hClient *HLINCLIENT, hHw HLINHW, pMsg *TLINMsg) error {
	return checkErr(procLINWrite.Call(
		uintptr(unsafe.Pointer(hClient)), 
		uintptr(hHw), 
		uintptr(unsafe.Pointer(pMsg))))
}

/*
	Initializes a Hardware with a given Mode and Baudrate.

	Remarks:
		If the Hardware was initialized by another Client, the function 
		will re-initialize the Hardware. All connected clients will be affected.
		It is the job of the user to manage the setting and/or configuration of 
		Hardware, e.g. by using the Boss-Client parameter of the Hardware.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_CLIENT, 
		TLIN_ERROR_ILLEGAL_HARDWARE, TLIN_ERROR_ILLEGAL_BAUDRATE

	Parameters:
		hClient     : Handle of the Client  (HLINCLIENT) 
		hHw         : Handle of the Hardware (HLINHW)
		byMode      : Hardware mode (see Hardware Operation Modes)
		wBaudrate   : Baudrate (see LIN_MIN_BAUDRATE and LIN_MAX_BAUDRATE) (c_ushort)

	Returns:
		A LIN Error Code
*/

func LIN_InitializeHardware(hClient *HLINCLIENT, hHw HLINHW, byMode WORD, wBaudrate WORD) error {
	return checkErr(procLINInitializeHardware.Call(
		uintptr(unsafe.Pointer(hClient)), 
		uintptr(hHw), 
		uintptr(byMode), 
		uintptr(wBaudrate)))
}

/*
	Gets an array containing the handles of the current Hardware available in the system.
	The count of Hardware handles returned in the array is written in 'pCount'.

	Remarks:
		To ONLY get the count of available Hardware, call this 
		function using 'pBuff' = NULL and wBuffSize = 0.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_BUFFER_INSUFFICIENT

	Parameters:
		pBuff       : Buffer for the handles (HLINHW[])
		wBuffSize   : Size of the buffer in bytes (pBuff.Length * 2) (c_ushort)
		pCount      : Number of Hardware available (c_ushort)

	Returns:
		A LIN Error Code
*/ 

func LIN_GetAvailableHardware(pBuff []HLINHW, wBuffSize WORD, pCount *WORD) error {
	ensureInit()

	return checkErr(procLINGetAvailableHardware.Call(
		uintptr(unsafe.Pointer(&pBuff[0])), 
		uintptr(wBuffSize), 
		uintptr(unsafe.Pointer(pCount))))
}

/*
	Sets a Hardware parameter to a given value.

	Remarks:
		Allowed TLINHardwareParam           Parameter
		values in wParam:                   type:       Description:
		-------------------------           ----------  ------------------------------
		TLIN_HARDWAREPARAM_MESSAGE_FILTER   c_uint64    Hardware message filter. Each bit
														corresponds to a Frame ID (0..63)
		TLIN_HARDWAREPARAM_BOSS_CLIENT      HLINCLIENT  Handle of the new Boss-Client
		TLIN_HARDWAREPARAM_ID_NUMBER        c_int       Identification number for a hardware
		TLIN_HARDWAREPARAM_USER_DATA        c_ubyte[]   User data to write on a hardware. See LIN_MAX_USER_DATA

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_TYPE, TLIN_ERROR_WRONG_PARAM_VALUE, 
		TLIN_ERROR_ILLEGAL_CLIENT, TLIN_ERROR_ILLEGAL_HARDWARE

	Parameters:
		hClient     : Handle of the Client  (HLINCLIENT) 
		hHw         : Handle of the Hardware (HLINHW)
		wParam      : Parameter to set (TLINHardwareParam)
		pBuff       : Parameter value buffer (see remarks)
		wBuffSize   : Value buffer size (ushort)

	Returns:
		A LIN Error Code
*/    
func LIN_SetHardwareParam(hClient *HLINCLIENT, hHw HLINHW, wParam TLINHardwareParam, pBuff *BYTE, wBuffSize WORD) error {
	return checkErr(procLINSetHardwareParam.Call(
		uintptr(unsafe.Pointer(hClient)),
		uintptr(hHw),
		uintptr(wParam),
		uintptr(unsafe.Pointer(pBuff)),
		uintptr(wBuffSize),
	))
}

/*
	Gets a Hardware parameter.

	Remarks:
		Allowed TLINHardwareParam               Parameter
		values in wParam:                       type:           Description:
		------------------------               ----------     ------------------------------
		TLIN_HARDWAREPARAM_NAME                 string          Name of the Hardware. See LIN_MAX_NAME_LENGTH
		TLIN_HARDWAREPARAM_DEVICE_NUMBER        c_int           Index of the Device owner of the Hardware
		TLIN_HARDWAREPARAM_CHANNEL_NUMBER       c_int           Channel Index of the Hardware on the owner device
		TLIN_HARDWAREPARAM_CONNECTED_CLIENTS    HLINCLIENT[]    Array of Client Handles conencted to a Hardware 
																The first item in the array refers to the 
																amount of handles. So [*] = Total handles + 1
		TLIN_HARDWAREPARAM_MESSAGE_FILTER       c_uint64        Configured message filter. Each bit corresponds 
																to a Frame ID (0..63)
		TLIN_HARDWAREPARAM_BAUDRATE             c_int           Configured baudrate
		TLIN_HARDWAREPARAM_MODE                 c_int           0 = Slave, otehrwise Master
		TLIN_HARDWAREPARAM_FIRMWARE_VERSION     TLINVersion     A TLINVersion structure containing the Firmware Version
		TLIN_HARDWAREPARAM_BUFFER_OVERRUN_COUNT c_int           Receive Buffer Overrun Counter
		TLIN_HARDWAREPARAM_BOSS_CLIENT          HLINCLIENT      Handle of the current Boss-Client
		TLIN_HARDWAREPARAM_SERIAL_NUMBER        c_int           Serial number of the Hardware
		TLIN_HARDWAREPARAM_VERSION              c_int           Version of the Hardware
		TLIN_HARDWAREPARAM_TYPE                 c_int           Type of the Hardware
		TLIN_HARDWAREPARAM_OVERRUN_COUNT        c_int           Receive Queue Buffer Overrun Counter
		TLIN_HARDWAREPARAM_ID_NUMBER            c_int           Identification number for a hardware
		TLIN_HARDWAREPARAM_USER_DATA            c_ubyte[]       User data saved on the hardware. See LIN_MAX_USER_DATA

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_TYPE, TLIN_ERROR_WRONG_PARAM_VALUE, 
		TLIN_ERROR_ILLEGAL_HARDWARE, TLIN_ERROR_BUFFER_INSUFFICIENT 

	Parameters:
		hHw         : Handle of the Hardware (HLINHW)
		wParam      : Parameter to get (TLINHardwareParam)
		pBuff       : Parameter buffer (see remarks)
		wBuffSize   : Buffer size (ushort)

	Returns:
		A LIN Error Code
*/

func LIN_GetHardwareParam(hHw HLINHW, wParam TLINHardwareParam, pBuff *BYTE, wBuffSize WORD) error {
	return checkErr(procLINGetHardwareParam.Call(
		uintptr(hHw),
		uintptr(wParam), 
		uintptr(unsafe.Pointer(pBuff)), 
		uintptr(wBuffSize)))
}

/*
	Flushes the queues of the Hardware and resets its counters.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED, 
		TLIN_ERROR_MANAGER_NOT_RESPONDING, 
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_CLIENT, 
		TLIN_ERROR_ILLEGAL_HARDWARE

	Parameters:
		hClient     : Handle of the Client  (HLINCLIENT) 
		hHw         : Handle of the Hardware (HLINHW)

	Returns:
		A LIN Error Code
*/    

func LIN_ResetHardware(hClient *HLINCLIENT, hHw HLINHW) error {
	return checkErr(procLINResetHardware.Call(
		uintptr(unsafe.Pointer(hClient)),
		uintptr(hHw)))
}			

/*
	Deletes the current configuration of the Hardware and sets its defaults.
	The Client 'hClient' must be registered and connected to the Hardware to 
	be accessed.
			
	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED, 
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_CLIENT, 
		TLIN_ERROR_ILLEGAL_HARDWARE

	Parameters:
		hClient     : Handle of the Client  (HLINCLIENT) 
		hHw         : Handle of the Hardware (HLINHW)

	Returns:
		A LIN Error Code
*/  

func LIN_ResetHardwareConfig(hClient *HLINCLIENT, hHw HLINHW) error {
	return checkErr(procLINResetHardwareConfig.Call(
		uintptr(unsafe.Pointer(hClient)),
		uintptr(hHw)))
}	

/*
	Physically identifies a LIN Hardware (a channel on a LIN Device) by 
	blinking its associated LED.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_HARDWARE

	Parameters:
		hHw         : Handle of the Hardware (HLINHW)

	Returns:
		A LIN Error Code
*/  

func LIN_IdentifyHardware(hHw HLINHW) error {
	return checkErr(procLINIdentifyHardware.Call(
		uintptr(hHw)))
}			


/*
	Modifies the filter of a Client and, eventually, the filter of the 
	connected Hardware. The messages with FrameID 'bFromFrameId' to 
	'bToFrameId' will be received.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_CLIENT, 
		TLIN_ERROR_ILLEGAL_HARDWARE, TLIN_ERROR_ILLEGAL_FRAMEID

	Parameters:
		hClient     : Handle of the Client  (HLINCLIENT) 
		hHw         : Handle of the Hardware (HLINHW)
		bFromFrameId: First ID of the frame range (c_ubyte)
		bToFrameId  : Last ID of the frame range (c_ubyte)

	Returns:
		A LIN Error Code
*/    

func LIN_RegisterFrameId(hClient *HLINCLIENT, hHw HLINHW, bFromFrameId BYTE, bToFrameId BYTE) error {
	return checkErr(procLINRegisterFrameId.Call(
		uintptr(unsafe.Pointer(hClient)), 
		uintptr(hHw), 
		uintptr(bFromFrameId), 
		uintptr(bToFrameId)))
}

/*
	Configures a LIN Frame in a given Hardware. The Client 'hClient' must 
	be registered and connected to the Hardware to be accessed.
			
	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_CLIENT, 
		TLIN_ERROR_ILLEGAL_HARDWARE, TLIN_ERROR_ILLEGAL_FRAMEID,
		TLIN_ERROR_ILLEGAL_LENGTH

	Parameters:
		hClient     : Handle of the Client  (HLINCLIENT) 
		hHw         : Handle of the Hardware (HLINHW)
		pFrameEntry : Frame entry information buffer (TLINFrameEntry)

	Returns:
		A LIN Error Code
*/   

func LIN_SetFrameEntry(hClient *HLINCLIENT, hHw HLINHW, pFrameEntry *TLINFrameEntry) error {
	return checkErr(procLINSetFrameEntry.Call(
		uintptr(unsafe.Pointer(hClient)), 
		uintptr(hHw), 
		uintptr(unsafe.Pointer(pFrameEntry))))
}    

/*
	Gets the configuration of a LIN Frame from a given Hardware.

	Remarks:
			The 'pFrameEntry.FrameId' must be set to the ID of the frame, whose 
			configuration should be returned.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_HARDWARE, 
		TLIN_ERROR_ILLEGAL_FRAMEID

	Parameters:
		hHw         : Handle of the Hardware (HLINHW)
		pFrameEntry : Frame entry buffer (TLINFrameEntry).
						The member FrameId  must be set to the ID of the frame, 
						whose configuration should  be returned

	Returns:
		A LIN Error Code
*/

func LIN_GetFrameEntry(hHw HLINHW, pFrameEntry *TLINFrameEntry) error {
	return checkErr(procLINGetFrameEntry.Call(
		uintptr(hHw), 
		uintptr(unsafe.Pointer(pFrameEntry))))
}
    
/*
	Updates the data of a LIN Frame for a given Hardware. The Client 
	'hClient' must be registered and connected to the Hardware to be 
	accessed. 'pData' must have at least a size of 'bLen'.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_CLIENT, 
		TLIN_ERROR_ILLEGAL_HARDWARE, TLIN_ERROR_ILLEGAL_FRAMEID, 
		TLIN_ERROR_ILLEGAL_LENGTH, TLIN_ERROR_ILLEGAL_INDEX, 
		TLIN_ERROR_ILLEGAL_RANGE

	Parameters:
		hClient     : Handle of the Client  (HLINCLIENT) 
		hHw         : Handle of the Hardware (HLINHW)
		bFrameId    : Frame ID (c_ubyte)
		bIndex      : Index where the update data Starts (0..7) (c_ubyte)
		bLen        : Count of Data bytes to be updated. (c_ubyte)
		pData       : Data buffer (c_ubyte[])

	Returns:
		A LIN Error Code
*/

func LIN_UpdateByteArray(hClient *HLINCLIENT, hHw HLINHW, bFrameId BYTE, bIndex BYTE, bLen BYTE, pData *BYTE) error {
	return checkErr(procLINUpdateByteArray.Call(
		uintptr(unsafe.Pointer(hClient)), 
		uintptr(hHw), 
		uintptr(bFrameId), 
		uintptr(bIndex), 
		uintptr(bLen), 
		uintptr(unsafe.Pointer(pData))))
}

/*
	Sets the Frame 'bFrameId' as Keep-Alive frame for the given Hardware and 
	starts to send it every 'wPeriod' Milliseconds. The Client 'hClient' must 
	be registered and connected to the Hardware to be accessed.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_CLIENT, 
		TLIN_ERROR_ILLEGAL_HARDWARE, TLIN_ERROR_ILLEGAL_FRAMEID,
		TLIN_ERROR_ILLEGAL_SCHEDULER_STATE, TLIN_ERROR_ILLEGAL_FRAME_CONFIGURATION

	Parameters:
		hClient     : Handle of the Client  (HLINCLIENT) 
		hHw         : Handle of the Hardware (HLINHW)
		bFrameId    : ID of the Keep-Alive Frame (c_ubyte)
		wPeriod     : Keep-Alive Interval in Milliseconds (c_ushort)

	Returns:
		A LIN Error Code
*/    
func LIN_StartKeepAlive(hClient *HLINCLIENT, hHw HLINHW, bFrameId BYTE, wPeriod WORD) error {
	return checkErr(procLINStartKeepAlive.Call(
		uintptr(unsafe.Pointer(hClient)), 
		uintptr(hHw),
		uintptr(bFrameId), 
		uintptr(wPeriod)))
}

/*
	Suspends the sending of a Keep-Alive frame in the given Hardware.
	The Client 'hClient' must be registered and connected to the Hardware 
	to be accessed.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_CLIENT, 
		TLIN_ERROR_ILLEGAL_HARDWARE

	Parameters:
		hClient     : Handle of the Client  (HLINCLIENT) 
		hHw         : Handle of the Hardware (HLINHW)

	Returns:
		A LIN Error Code
*/
			
func LIN_SuspendKeepAlive(hClient *HLINCLIENT, hHw HLINHW) error {
	return checkErr(procLINSuspendKeepAlive.Call(
		uintptr(unsafe.Pointer(hClient)), 
		uintptr(hHw)))
}

/*
	Resumes the sending of a Keep-Alive frame in the given Hardware. 
	The Client 'hClient' must be registered and connected to the Hardware 
	to be accessed.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_CLIENT, 
		TLIN_ERROR_ILLEGAL_HARDWARE, TLIN_ERROR_ILLEGAL_SCHEDULER_STATE, 
		TLIN_ERROR_ILLEGAL_FRAME_CONFIGURATION

	Parameters:
		hClient     : Handle of the Client  (HLINCLIENT) 
		hHw         : Handle of the Hardware (HLINHW)

	Returns:
		A LIN Error Code
*/

func LIN_ResumeKeepAlive(hClient *HLINCLIENT, hHw HLINHW) error {
	return checkErr(procLINResumeKeepAlive.Call(
		uintptr(unsafe.Pointer(hClient)), 
		uintptr(hHw)))
}

/*
	Configures the slots of a Schedule in a given Hardware. 

	Remarks:
		The Client 'hClient' must be registered and connected to the Hardware
		to be accessed. The Slot handles will be returned in the parameter 
		"pSchedule" (Slots buffer), when this function successfully completes. 

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_CLIENT, 
		TLIN_ERROR_ILLEGAL_HARDWARE, 
		TLIN_ERROR_ILLEGAL_SCHEDULENUMBER, 
		TLIN_ERROR_ILLEGAL_SLOTCOUNT,
		TLIN_ERROR_SCHEDULE_SLOT_POOL_FULL

	Parameters:
		hClient         : Handle of the Client  (HLINCLIENT) 
		hHw             : Handle of the Hardware (HLINHW)
		iScheduleNumber : Schedule number (c_int)
							(see LIN_MIN_SCHEDULE_NUMBER and LIN_MAX_SCHEDULE_NUMBER)
		pSchedule       : Slots buffer (TLINScheduleSlot[])
		iSlotCount      : Count of Slots in the slots buffer (c_int)

	Returns:
		A LIN Error Code
*/

func LIN_SetSchedule(hClient *HLINCLIENT, hHw HLINHW, iScheduleNumber SDWORD, pSchedule *TLINScheduleSlot, iSlotCount SDWORD) error {
	return checkErr(procLINSetSchedule.Call(
		uintptr(unsafe.Pointer(hClient)), 
		uintptr(hHw), 
		uintptr(iScheduleNumber), 
		uintptr(unsafe.Pointer(pSchedule)), 
		uintptr(iSlotCount)))
}

/*
	Gets the slots of a Schedule from a given Hardware. The count of slots 
	returned in the array is written in 'pSlotCount'.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_HARDWARE, 
		TLIN_ERROR_ILLEGAL_SCHEDULENUMBER, 
		TLIN_ERROR_ILLEGAL_SLOTCOUNT,
		TLIN_ERROR_ILLEGAL_SCHEDULE

	Parameters:
		hHw             : Handle of the Hardware (HLINHW)
		iScheduleNumber : Schedule Number (see LIN_MIN_SCHEDULE_NUMBER and LIN_MAX_SCHEDULE_NUMBER) //THIS IS SDWORD per the variable names in PLinTypes.go
		pScheduleBuff   : Slots Buffer (TLINScheduleSlot[])
		iMaxSlotCount   : Maximum number of slots to read. (c_int)
		pSlotCount      : Real number of slots read. (c_int)

	Returns:
		A LIN Error Code
*/

func LIN_GetSchedule(hHw HLINHW, iScheduleNumber SDWORD, pScheduleBuff *TLINScheduleSlot, iMaxSlotCount SDWORD, pSlotCount *SDWORD) error {
	return checkErr(procLINGetSchedule.Call(
		uintptr(hHw), 
		uintptr(iScheduleNumber), 
		uintptr(unsafe.Pointer(pScheduleBuff)), 
		uintptr(iMaxSlotCount), 
		uintptr(unsafe.Pointer(pSlotCount))))
}

/*
	Removes all slots contained by a Schedule of a given Hardware. The 
	Client 'hClient' must be registered and connected to the Hardware to 
	be accessed.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_CLIENT, 
		TLIN_ERROR_ILLEGAL_HARDWARE, 
		TLIN_ERROR_ILLEGAL_SCHEDULENUMBER,
		TLIN_ERROR_ILLEGAL_SCHEDULER_STATE

	Parameters:
		hClient     : Handle of the Client  (HLINCLIENT) 
		hHw         : Handle of the Hardware (HLINHW)
		iScheduleNumber : Schedule Number (c_int)
							(see LIN_MIN_SCHEDULE_NUMBER and LIN_MAX_SCHEDULE_NUMBER)

	Returns:
		A LIN Error Code
*/

func LIN_DeleteSchedule(hClient *HLINCLIENT, hHw HLINHW, iScheduleNumber SDWORD) error {
	return checkErr(procLINDeleteSchedule.Call(
		uintptr(unsafe.Pointer(hClient)), 
		uintptr(hHw), 
		uintptr(iScheduleNumber)))
}


/*
	Sets a 'breakpoint' on a slot from a Schedule in a given Hardware. The 
	Client 'hClient' must be registered and connected to the Hardware to 
	be accessed.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_CLIENT, 
		TLIN_ERROR_ILLEGAL_HARDWARE

	Parameters:
		hClient             : Handle of the Client  (HLINCLIENT) 
		hHw                 : Handle of the Hardware (HLINHW)
		iBreakPointNumber   : Breakpoint Number (0 or 1) (c_int)
		dwHandle            : Slot Handle (c_uint)

	Returns:
		A LIN Error Code
*/

func LIN_SetScheduleBreakPoint(hClient *HLINCLIENT, hHw HLINHW, iBreakPointNumber SDWORD, dwHandle DWORD) error {
	return checkErr(procLINSetScheduleBreakPoint.Call(
		uintptr(unsafe.Pointer(hClient)), 
		uintptr(hHw), 
		uintptr(iBreakPointNumber), 
		uintptr(dwHandle)))
}			

/*
	Activates a Schedule in a given Hardware. The Client 'hClient' must 
	be registered and connected to the Hardware to be accessed.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_CLIENT, 
		TLIN_ERROR_ILLEGAL_HARDWARE, 
		TLIN_ERROR_ILLEGAL_SCHEDULENUMBER,
		TLIN_ERROR_ILLEGAL_SCHEDULE,
		TLIN_ERROR_ILLEGAL_HARDWARE_MODE

	Parameters:
		hClient         : Handle of the Client  (HLINCLIENT) 
		hHw             : Handle of the Hardware (HLINHW)
		iScheduleNumber : Schedule Number (c_int/int32)
							(see LIN_MIN_SCHEDULE_NUMBER and LIN_MAX_SCHEDULE_NUMBER)

	Returns:
		A LIN Error Code
*/

func LIN_StartSchedule(hClient *HLINCLIENT, hHw HLINHW, iScheduleNumber SDWORD) error {
	return checkErr(procLINStartSchedule.Call(
		uintptr(unsafe.Pointer(hClient)), 
		uintptr(hHw), 
		uintptr(iScheduleNumber)))
}

/*
	Suspends an active Schedule in a given Hardware. The Client 'hClient' 
	must be registered and connected to the Hardware to be accessed.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_CLIENT,
		TLIN_ERROR_ILLEGAL_HARDWARE

	Parameters:
		hClient     : Handle of the Client  (HLINCLIENT) 
		hHw         : Handle of the Hardware (HLINHW)

	Returns:
		A LIN Error Code
*/

func LIN_SuspendSchedule(hClient *HLINCLIENT, hHw HLINHW) error {
	return checkErr(procLINSuspendSchedule.Call(
		uintptr(unsafe.Pointer(hClient)), 
		uintptr(hHw)))
}

/*
	Restarts a configured Schedule in a given Hardware. The Client 'hClient' 
	must be registered and connected to the Hardware to be accessed.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED, 
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_CLIENT,
		TLIN_ERROR_ILLEGAL_HARDWARE, TLIN_ERROR_ILLEGAL_SCHEDULE,
		TLIN_ERROR_ILLEGAL_HARDWARE_MODE, 
		TLIN_ERROR_ILLEGAL_SCHEDULER_STATE

	Parameters:
		hClient     : Handle of the Client  (HLINCLIENT) 
		hHw         : Handle of the Hardware (HLINHW)

	Returns:
		A LIN Error Code
*/

func LIN_ResumeSchedule(hClient *HLINCLIENT, hHw HLINHW) error {
	return checkErr(procLINResumeSchedule.Call(
		uintptr(unsafe.Pointer(hClient)), 
		uintptr(hHw)))
}

/*
	Sends a wake-up message impulse (single data byte 0xF0). The Client 
	'hClient' must be registered and connected to the Hardware to be 
	accessed.

	Remark: Only in Slave-mode. After sending a wake-up impulse a time
	of 150 milliseconds is used as timeout.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED, 
		TLIN_ERROR_MANAGER_NOT_RESPONDING, 
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_CLIENT,
		TLIN_ERROR_ILLEGAL_HARDWARE

	Parameters:
		hClient     : Handle of the Client  (HLINCLIENT) 
		hHw         : Handle of the Hardware (HLINHW)

	Returns:
		A LIN Error Code
*/

func LIN_XmtWakeUp(hClient *HLINCLIENT, hHw HLINHW) error {
	return checkErr(procLINXmtWakeUp.Call(
		uintptr(unsafe.Pointer(hClient)), 
		uintptr(hHw)))
}

/*
	Sends a wake-up message impulse (single data byte 0xF0) and specifies
	a custom bus-sleep timeout, in milliseconds. The Client 'hClient' 
	must be registered and connected to the Hardware to be accessed.

	Remark: Only in Slave-mode. The bus-sleep timeout is set to its  
	default, 150 milliseconds, after the custom timeout is exhausted.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED, 
		TLIN_ERROR_MANAGER_NOT_RESPONDING, 
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_CLIENT,
		TLIN_ERROR_ILLEGAL_HARDWARE

	Parameters:
		hClient     : Handle of the Client  (HLINCLIENT) 
		hHw         : Handle of the Hardware (HLINHW)
		sTimeOut    : Bus-sleep timeout (c_ushort)

	Returns:
		A LIN Error Code
*/

func LIN_XmtDynamicWakeUp(hClient *HLINCLIENT, hHw HLINHW, sTimeOut WORD) error {
	return checkErr(procLINXmtDynamicWakeUp.Call(
		uintptr(unsafe.Pointer(hClient)), 
		uintptr(hHw), 
		uintptr(sTimeOut)))
}

/*
	Starts a process to detect the Baud rate of the LIN bus that is 
	connected to the indicated Hardware.
	The Client 'hClient' must be registered and connected to the Hardware 
	to be accessed. The Hardware must be not initialized in order 
	to do an Auto-baudrate procedure.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_CLIENT,
		TLIN_ERROR_ILLEGAL_HARDWARE, TLIN_ERROR_ILLEGAL_HARDWARE_STATE

	Parameters:
		hClient     : Handle of the Client  (HLINCLIENT) 
		hHw         : Handle of the Hardware (HLINHW)
		wTimeOut    : Auto-baudrate Timeout in Milliseconds (c_ushort, uint16)

	Returns:
		A LIN Error Code
*/

func LIN_StartAutoBaud(hClient *HLINCLIENT, hHw HLINHW, wTimeOut WORD) error {
	return checkErr(procLINStartAutoBaud.Call(
		uintptr(unsafe.Pointer(hClient)), 
		uintptr(hHw), 
		uintptr(wTimeOut)))
}

/*
	Retrieves current status information from the given Hardware.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_HARDWARE

	Parameters:
		hHw         : Handle of the Hardware (HLINHW)
		pStatusBuff : Status data buffer (TLINHardwareStatus)

	Returns:
		A LIN Error Code
*/
func LIN_GetStatus(hHw HLINHW, pStatusBuff *TLINHardwareStatus) error {
	return checkErr(procLINGetStatus.Call(
		uintptr(hHw), 
		uintptr(unsafe.Pointer(pStatusBuff))))
}

/*
	Calculates the checksum of a LIN Message and writes it into the 
	'Checksum' field of 'pMsg'.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING, 
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_LENGTH        

	Parameters:
		pMsg        : Message buffer (TLINMsg)

	Returns:
		A LIN Error Code
*/

func LIN_CalculateChecksum(pMsg *TLINMsg) error {
	return checkErr(procLINCalculateChecksum.Call(
		uintptr(unsafe.Pointer(pMsg))))
}

/*
	Returns a TLINVersion structure containing the PLIN-API DLL version.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE

	Parameters:
		pVerBuffer  : Version buffer (TLINVersion)

	Returns:
		A LIN Error Code
*/

func LIN_GetVersion(pVerBuffer *TLINVersion) error {
	return checkErr(procLINGetVersion.Call(
		uintptr(unsafe.Pointer(pVerBuffer))))
}

/*
	Returns a string containing Copyright information.
	
	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE
	
	Parameters:
		pTextBuff   : String buffer (character array from create_string_buffer)
		wBuffSize   : Size in bytes of the buffer (c_int)

	Returns:
		A LIN Error Code
*/

func LIN_GetVersionInfo(pTextBuff LPSTR, wBuffSize WORD) error {
	ensureInit()

	return checkErr(procLINGetVersionInfo.Call(
		uintptr(unsafe.Pointer(pTextBuff)),
		uintptr(wBuffSize)))
}
//////////////////////////
/*
	Converts the error code 'dwError' to a text containing an error 
	description in the language given as parameter (when available).
	
	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED, 
		TLIN_ERROR_MANAGER_NOT_RESPONDING, 
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_BUFFER_INSUFFICIENT
	
	Parameters:
		dwError     : A TLINError Code (TLINError)
		bLanguage   : Indicates a "Primary language ID" (c_ubyte)
		strTextBuff : Error string buffer (character array from create_string_buffer)
		wBuffSize   : Buffer size in bytes (c_int)

	Returns:
		A LIN Error Code
*/

func LIN_GetErrorText(dwError TLINError, bLanguage BYTE, strTextBuff LPSTR, wBuffSize WORD) error {
	ensureInit()

	return checkErr(procLINGetErrorText.Call(
		uintptr(dwError), 
		uintptr(bLanguage), 
		uintptr(unsafe.Pointer(strTextBuff)), 
		uintptr(wBuffSize)))
}

/*
	Gets the 'FrameId with Parity' corresponding to the given 
	'pFrameId' and writes the result on it.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED, 
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_FRAMEID

	Parameters:
		pframeid    : Frame ID (0..LIN_MAX_FRAME_ID) (c_ubyte)

	Returns:
		A LIN Error Code
*/
/*func LIN_GetPID(pframeid *BYTE) error {
	return checkErr(procLINGetPID.Call(
		uintptr(unsafe.Pointer(pframeid))))
}*/
func LIN_GetPID(pid *BYTE) error {

	ensureInit()

	return checkErr(procLINGetPID.Call(
		uintptr(unsafe.Pointer(pid)),
	))
}

/*
	Gets the system time used by the LIN-USB adapter.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_HARDWARE

	Parameters:
		hHw         : Handle of the Hardware (HLINHW)
		pTargetTime : Target Time buffer (c_uint64)

	Returns:
		A LIN Error Code
*/

func LIN_GetTargetTime(hHw HLINHW, pTargetTime *UINT64) error {
	return checkErr(procLINGetTargetTime.Call(
		uintptr(hHw), 
		uintptr(unsafe.Pointer(pTargetTime))))
}

/*
	Sets the Response Remap of a LIN Slave.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING, 
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_FRAMEID, 
		TLIN_ERROR_ILLEGAL_CLIENT, TLIN_ERROR_ILLEGAL_HARDWARE,
		TLIN_ERROR_MEMORY_ACCESS

	Parameters:
		hClient     : Handle of the Client  (HLINCLIENT) 
		hHw         : Handle of the Hardware (HLINHW)
		pRemapTab   : Remap Response buffer (c_ubyte[64])

	Returns:
		A LIN Error Code
*/

func LIN_SetResponseRemap(hClient *HLINCLIENT, hHw HLINHW, pRemapTab *[64]BYTE) error {
	return checkErr(procLINSetResponseRemap.Call(
		uintptr(unsafe.Pointer(hClient)), 
		uintptr(hHw), 
		uintptr(unsafe.Pointer(pRemapTab))))
}


/*
	Gets the Response Remap of a LIN Slave.

	Possible DLL interaction errors:
		TLIN_ERROR_MANAGER_NOT_LOADED,
		TLIN_ERROR_MANAGER_NOT_RESPONDING,
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE, TLIN_ERROR_ILLEGAL_CLIENT,
		TLIN_ERROR_ILLEGAL_HARDWARE

	Parameters:
		hHw         : Handle of the Hardware (HLINHW)
		pRemapTab   : Remap Response buffer (c_ubyte[64])

	Returns:
		A LIN Error Code
*/
func LIN_GetResponseRemap(hHw HLINHW, pRemapTab *[64]BYTE) error {
	return checkErr(procLINGetResponseRemap.Call(
		uintptr(hHw), 
		uintptr(unsafe.Pointer(pRemapTab))))
}

/*
	Gets the current system time. The system time is returned by 
	Windows as the elapsed number of microseconds since system start.

	Possible DLL interaction errors:
		TLIN_ERROR_MEMORY_ACCESS

	Possible API errors:
		TLIN_ERROR_WRONG_PARAM_VALUE

	Parameters:
		pTargetTime : System Time buffer (c_uint64)

	Returns:
		A LIN Error Code
*/

func LIN_GetSystemTime(pTargetTime *UINT64) error {
	return checkErr(procLINGetSystemTime.Call(
		uintptr(unsafe.Pointer(pTargetTime))))
}



