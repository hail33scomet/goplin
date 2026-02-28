package plin

/* -----------------------------------------------------------------------------
 Base type mappings (C → Go)
 -----------------------------------------------------------------------------
*/
// C: BYTE   -> uint8
// C: WORD   -> uint16
// C: DWORD  -> uint32
// C: UINT64 -> uint64
// C: LPSTR  -> char* (C string)
// We'll model LPSTR as uintptr for now (pointer-sized). You can change to *C.char in cgo builds.

type BYTE = uint8
type WORD = uint16
type DWORD = uint32
type SDWORD = int32
type UINT64 = uint64
type LPSTR = uintptr // or *C.char in a cgo context

// -----------------------------------------------------------------------------
// Type definitions from header typedefs
// -----------------------------------------------------------------------------

type HLINCLIENT uint8             // LIN client handle
type HLINHW uint16                // LIN hardware handle
type TLINMsgErrors uint32          // Error flags for LIN Rcv Msgs
type TLINClientParam uint16       // Client Parameters (GetClientParam Function)
type TLINHardwareParam uint16     // Hardware Parameters (GetHardwareParam function)
type TLINMsgType uint8            // Received Message Types
type TLINSlotType uint8           // Schedule Slot Types
type TLINDirection uint8          // Message Direction Types
type TLINChecksumType uint8       // Message Checksum Types
type TLINHardwareMode uint8       // Hardware Operation Modes
type TLINHardwareState uint8      // Hardware Status
type TLINScheduleState uint8      // Schedule Status
type TLINError int32             // Error Codes


/////////////////////////////////////////////////////////////
// Value definitions
/////////////////////////////////////////////////////////////

// Invalid Handle values
const (
	INVALID_LIN_HANDLE = BYTE(0)         // Invalid value for all LIN handles (Client, Hardware)
                                            
	// Hardware Types                                                                 
	LIN_HW_TYPE_USB = BYTE(1)            // LIN USB type
	LIN_HW_TYPE_USB_PRO = BYTE(1)        // PCAN-USB Pro LIN type
	LIN_HW_TYPE_USB_PRO_FD = BYTE(2)     // PCAN-USB Pro FD LIN type
	LIN_HW_TYPE_PLIN_USB = BYTE(3)       // PLIN-USB type
)
                                            
// Minimum and Maximum values   
const (                                                       
	LIN_MAX_FRAME_ID = BYTE(63)       // Maximum allowed Frame ID (0x3F)
	LIN_MAX_SCHEDULES = SDWORD(8)          // Maximum allowed Schedules per Hardware
	LIN_MIN_SCHEDULE_NUMBER = SDWORD(0)          // Minimum Schedule number
	LIN_MAX_SCHEDULE_NUMBER = SDWORD(7)          // Maximum Schedule number
	LIN_MAX_SCHEDULE_SLOTS = SDWORD(256)        // Maximum allowed Schedule slots per Hardware
	LIN_MIN_BAUDRATE = WORD(1000)    // Minimum LIN Baudrate
	LIN_MAX_BAUDRATE = WORD(20000)   // Maximum LIN Baudrate
	LIN_MAX_NAME_LENGTH = WORD(48)      // Maximum number of bytes for Name / ID of a Hardware or Client
	LIN_MAX_USER_DATA = SDWORD(24)         // Maximum number of bytes that a user can read/write on a Hardware
	LIN_MIN_BREAK_LENGTH = SDWORD(13)         // Minimum number of bits that can be used as break field in a LIN frame
	LIN_MAX_BREAK_LENGTH = SDWORD(32)         // Maximum number of bits that can be used as break field in a LIN frame
	LIN_MAX_RCV_QUEUE_COUNT = SDWORD(65535)      // Maximum number of LIN frames that can be stored in the reception queue of a client
)

// Frame flags for LIN Frame Entries
const (
	FRAME_FLAG_RESPONSE_ENABLE = WORD(0x1)       // Slave Enable Publisher Response
	FRAME_FLAG_SINGLE_SHOT = WORD(0x2)           // Slave Publisher Single shot
	FRAME_FLAG_IGNORE_INIT_DATA = WORD(0x4)     // Ignore InitialData on set frame entry
)

// Flags for information in debug logs
const (
	LOG_FLAG_DEFAULT = WORD(0x0)     // Logs system exceptions / errors
	LOG_FLAG_ENTRY = WORD(0x1)     // Logs the entries to the PLIN-API functions 
	LOG_FLAG_PARAMETERS = WORD(0x2)     // Logs the parameters passed to the PLIN-API functions 
	LOG_FLAG_LEAVE = WORD(0x4)     // Logs the exits from the PLIN-API functions 
	LOG_FLAG_WRITE = WORD(0x8)     // Logs the LIN messages passed to the LIN_Write function
	LOG_FLAG_READ = WORD(0x10)    // Logs the LIN messages received within the LIN_Read function
	LOG_FLAG_ALL = WORD(0xFFFF)  // Logs all possible information within the PLIN-API functions
)

// Error flags for LIN Rcv Msgs
const (
	TLIN_MSGERROR_OK                     TLINError = 0x000  // No error
	TLIN_MSGERROR_INCONSISTENT_SYNCH     TLINError = 0x001  // Error on Synchronization field
	TLIN_MSGERROR_ID_PARITY_BIT_0        TLINError = 0x002  // Wrong parity Bit 0
	TLIN_MSGERROR_ID_PARITY_BIT_1        TLINError = 0x004  // Wrong parity Bit 1
	TLIN_MSGERROR_SLAVE_NOT_RESPONDING   TLINError = 0x008  // Slave not responding error
	TLIN_MSGERROR_TIMEOUT                TLINError = 0x010  // A timeout was reached
	TLIN_MSGERROR_CHECKSUM               TLINError = 0x020  // Wrong checksum
	TLIN_MSGERROR_GROUND_SHORT           TLINError = 0x040  // Bus shorted to ground
	TLIN_MSGERROR_VBAT_SHORT             TLINError = 0x080  // Bus shorted to Vbat
	TLIN_MSGERROR_SLOT_DELAY             TLINError = 0x100  // A slot time (delay) was too small
	TLIN_MSGERROR_OTHER_RESPONSE         TLINError = 0x200  // Response was received from other station
)

// Client Parameters (GetClientParam Function)
const (
	TLIN_CLIENTPARAM_NAME                    TLINClientParam = 1   // Client Name
	TLIN_CLIENTPARAM_MESSAGE_ON_QUEUE        TLINClientParam = 2   // Unread messages in the Receive Queue
	TLIN_CLIENTPARAM_WINDOW_HANDLE           TLINClientParam = 3   // Registered windows handle (information purpose)
	TLIN_CLIENTPARAM_CONNECTED_HARDWARE      TLINClientParam = 4   // Handles of the connected Hardware
	TLIN_CLIENTPARAM_TRANSMITTED_MESSAGES    TLINClientParam = 5   // Number of transmitted messages
	TLIN_CLIENTPARAM_RECEIVED_MESSAGES       TLINClientParam = 6   // Number of received messages
	TLIN_CLIENTPARAM_RECEIVE_STATUS_FRAME    TLINClientParam = 7   // Status of the property "Status Frames"
	TLIN_CLIENTPARAM_ON_RECEIVE_EVENT_HANDLE TLINClientParam = 8   // Handle of the Receive event
	TLIN_CLIENTPARAM_ON_PLUGIN_EVENT_HANDLE  TLINClientParam = 9   // Handle of the Hardware plug-in event
	TLIN_CLIENTPARAM_LOG_STATUS              TLINClientParam = 10  // Debug-Log activation status
	TLIN_CLIENTPARAM_LOG_CONFIGURATION       TLINClientParam = 11  // Configuration of the debugged information
)

// Hardware Parameters (GetHardwareParam function)
const (
	TLIN_HARDWAREPARAM_NAME                    TLINHardwareParam = 1   // Hardware / Device Name
	TLIN_HARDWAREPARAM_DEVICE_NUMBER           TLINHardwareParam = 2   // Index of the owner Device
	TLIN_HARDWAREPARAM_CHANNEL_NUMBER          TLINHardwareParam = 3   // Channel Index on the owner device
	TLIN_HARDWAREPARAM_CONNECTED_CLIENTS       TLINHardwareParam = 4   // Handles of the connected clients
	TLIN_HARDWAREPARAM_MESSAGE_FILTER          TLINHardwareParam = 5   // Message filter
	TLIN_HARDWAREPARAM_BAUDRATE                TLINHardwareParam = 6   // Baudrate
	TLIN_HARDWAREPARAM_MODE                    TLINHardwareParam = 7   // Master status
	TLIN_HARDWAREPARAM_FIRMWARE_VERSION        TLINHardwareParam = 8   // LIN hardware firmware version
	TLIN_HARDWAREPARAM_BUFFER_OVERRUN_COUNT    TLINHardwareParam = 9   // Receive Buffer Overrun Counter
	TLIN_HARDWAREPARAM_BOSS_CLIENT             TLINHardwareParam = 10  // Registered master Client
	TLIN_HARDWAREPARAM_SERIAL_NUMBER           TLINHardwareParam = 11  // Serial number of a Hardware
	TLIN_HARDWAREPARAM_VERSION                 TLINHardwareParam = 12  // Version of a Hardware
	TLIN_HARDWAREPARAM_TYPE                    TLINHardwareParam = 13  // Type of a Hardware
	TLIN_HARDWAREPARAM_OVERRUN_COUNT           TLINHardwareParam = 14  // Receive Queue Buffer Overrun Counter
	TLIN_HARDWAREPARAM_ID_NUMBER               TLINHardwareParam = 15  // Hardware identification number
	TLIN_HARDWAREPARAM_USER_DATA               TLINHardwareParam = 16  // User data on a hardware
	TLIN_HARDWAREPARAM_BREAK_LENGTH            TLINHardwareParam = 17  // Number of bits used as break field
	TLIN_HARDWAREPARAM_LIN_TERMINATION          TLINHardwareParam = 18  // LIN Termination status
	TLIN_HARDWAREPARAM_FLASH_MODE              TLINHardwareParam = 19  // Device flash mode
	TLIN_HARDWAREPARAM_SCHEDULE_ACTIVE         TLINHardwareParam = 20  // Active schedule number
	TLIN_HARDWAREPARAM_SCHEDULE_STATE          TLINHardwareParam = 21  // Schedule operation state
	TLIN_HARDWAREPARAM_SCHEDULE_SUSPENDED_SLOT TLINHardwareParam = 22  // Suspended schedule slot
)

// Received Message Types
const (
	TLIN_MSGTYPE_STANDARD             TLINMsgType = 0  // Standard LIN Message
	TLIN_MSGTYPE_BUS_SLEEP            TLINMsgType = 1  // Bus Sleep status message
	TLIN_MSGTYPE_BUS_WAKEUP           TLINMsgType = 2  // Bus WakeUp status message
	TLIN_MSGTYPE_AUTOBAUDRATE_TIMEOUT TLINMsgType = 3  // Auto-baudrate Timeout
	TLIN_MSGTYPE_AUTOBAUDRATE_REPLY   TLINMsgType = 4  // Auto-baudrate Reply
	TLIN_MSGTYPE_OVERRUN              TLINMsgType = 5  // Bus Overrun status
	TLIN_MSGTYPE_QUEUE_OVERRUN        TLINMsgType = 6  // Queue Overrun status
	TLIN_MSGTYPE_CLIENT_QUEUE_OVERRUN TLINMsgType = 7  // Client queue overrun
)

// Schedule Slot Types
const (
	TLIN_SLOTTYPE_UNCONDITIONAL  TLINSlotType = 0  // Unconditional frame
	TLIN_SLOTTYPE_EVENT          TLINSlotType = 1  // Event frame
	TLIN_SLOTTYPE_SPORADIC       TLINSlotType = 2  // Sporadic frame
	TLIN_SLOTTYPE_MASTER_REQUEST TLINSlotType = 3  // Diagnostic Master Request
	TLIN_SLOTTYPE_SLAVE_RESPONSE TLINSlotType = 4  // Diagnostic Slave Response
)

// Message Direction Types
const (
	TLIN_DIRECTION_DISABLED              TLINDirection = 0  // Direction disabled
	TLIN_DIRECTION_PUBLISHER             TLINDirection = 1  // Publisher
	TLIN_DIRECTION_SUBSCRIBER            TLINDirection = 2  // Subscriber
	TLIN_DIRECTION_SUBSCRIBER_AUTOLENGTH TLINDirection = 3  // Subscriber (auto length)
)

// Message Checksum Types
const (
	TLIN_CHECKSUMTYPE_CUSTOM   TLINChecksumType = 0  // Custom checksum
	TLIN_CHECKSUMTYPE_CLASSIC  TLINChecksumType = 1  // Classic checksum
	TLIN_CHECKSUMTYPE_ENHANCED TLINChecksumType = 2  // Enhanced checksum
	TLIN_CHECKSUMTYPE_AUTO     TLINChecksumType = 3  // Auto detect checksum
)

// Hardware Operation Modes
const (
	TLIN_HARDWAREMODE_NONE   TLINHardwareMode = 0  // Not initialized
	TLIN_HARDWAREMODE_SLAVE  TLINHardwareMode = 1  // Slave mode
	TLIN_HARDWAREMODE_MASTER TLINHardwareMode = 2  // Master mode

	// Hardware Status
	TLIN_HARDWARESTATE_NOT_INITIALIZED TLINHardwareState = 0  // Not initialized
	TLIN_HARDWARESTATE_AUTOBAUDRATE    TLINHardwareState = 1  // Detecting baudrate
	TLIN_HARDWARESTATE_ACTIVE          TLINHardwareState = 2  // Active
	TLIN_HARDWARESTATE_SLEEP           TLINHardwareState = 3  // Sleep
	TLIN_HARDWARESTATE_SHORT_GROUND    TLINHardwareState = 6  // Short to ground
	TLIN_HARDWARESTATE_VBAT_MISSING    TLINHardwareState = 7  // VBAT missing
)

// Schedule Status
const (
	TLIN_SCHEDULESTATE_NOT_RUNNING TLINScheduleState = 0  // Not running
	TLIN_SCHEDULESTATE_SUSPENDED   TLINScheduleState = 1  // Suspended
	TLIN_SCHEDULESTATE_RUNNING    TLINScheduleState = 2  // Running
)

// Error Codes
const (
	TLIN_ERROR_OK                          TLINError = 0      // Success
	TLIN_ERROR_XMTQUEUE_FULL               TLINError = 1      // Transmit Queue full
	TLIN_ERROR_ILLEGAL_PERIOD              TLINError = 2      // Invalid period
	TLIN_ERROR_RCVQUEUE_EMPTY              TLINError = 3      // Receive Queue empty
	TLIN_ERROR_ILLEGAL_CHECKSUMTYPE        TLINError = 4      // Invalid checksum type
	TLIN_ERROR_ILLEGAL_HARDWARE            TLINError = 5      // Invalid hardware handle
	TLIN_ERROR_ILLEGAL_CLIENT              TLINError = 6      // Invalid client handle
	TLIN_ERROR_WRONG_PARAM_TYPE            TLINError = 7      // Invalid parameter type
	TLIN_ERROR_WRONG_PARAM_VALUE           TLINError = 8      // Invalid parameter value
	TLIN_ERROR_ILLEGAL_DIRECTION           TLINError = 9      // Invalid direction
	TLIN_ERROR_ILLEGAL_LENGTH              TLINError = 10     // Invalid length
	TLIN_ERROR_ILLEGAL_BAUDRATE            TLINError = 11     // Invalid baudrate
	TLIN_ERROR_ILLEGAL_FRAMEID             TLINError = 12     // Invalid frame ID
	TLIN_ERROR_BUFFER_INSUFFICIENT         TLINError = 13     // Buffer too small
	TLIN_ERROR_ILLEGAL_SCHEDULENUMBER      TLINError = 14     // Invalid schedule number
	TLIN_ERROR_ILLEGAL_SLOTCOUNT           TLINError = 15     // Invalid slot count
	TLIN_ERROR_ILLEGAL_INDEX               TLINError = 16     // Invalid index
	TLIN_ERROR_ILLEGAL_RANGE               TLINError = 17     // Invalid range
	TLIN_ERROR_ILLEGAL_HARDWARE_STATE      TLINError = 18     // Invalid hardware state
	TLIN_ERROR_ILLEGAL_SCHEDULER_STATE     TLINError = 19     // Invalid scheduler state
	TLIN_ERROR_ILLEGAL_FRAME_CONFIGURATION TLINError = 20     // Invalid frame configuration
	TLIN_ERROR_SCHEDULE_SLOT_POOL_FULL     TLINError = 21     // Slot pool full
	TLIN_ERROR_ILLEGAL_SCHEDULE            TLINError = 22     // No schedule present
	TLIN_ERROR_ILLEGAL_HARDWARE_MODE       TLINError = 23     // Invalid hardware mode
	TLIN_ERROR_OUT_OF_RESOURCE             TLINError = 1001   // Out of resources
	TLIN_ERROR_MANAGER_NOT_LOADED          TLINError = 1002   // Manager not running
	TLIN_ERROR_MANAGER_NOT_RESPONDING      TLINError = 1003   // Manager not responding
	TLIN_ERROR_MEMORY_ACCESS               TLINError = 1004   // Memory access violation
	TLIN_ERROR_NOT_IMPLEMENTED             TLINError = 0xFFFE // Not implemented
	TLIN_ERROR_UNKNOWN                     TLINError = 0xFFFF // Unknown error
)


// Represents a Version Information
type TLINVersion struct {
	// Size = 8 bytes
	Major WORD //0 +0  Major part of a version number
	Minor WORD //1 +2  Minor part of a version number
	Revision WORD //2 +4  Revision part of a version number
	Build WORD //3 +6  Build part of a version number
}

// Represents a LIN Message to be sent
type TLINMsg struct {
	// Size = 13 bytes
	FrameId BYTE //0 +0  Frame ID (6 bit) + Parity (2 bit)
	Length BYTE //1 +1  Frame Length (1..8)
	Direction TLINDirection //2 +2  Frame Direction (see Message Direction Types)
	ChecksumType TLINChecksumType //3 +3  Frame Checksum type (see Message Checksum Types)
	Data [8]BYTE //4 +4  Data bytes (0..7)
	Checksum BYTE //5 +12 Frame Checksum
}

// Represents a received LIN Message
type TLINRcvMsg struct {
	// Size = 40 bytes	
	Type TLINMsgType //0 +0  Frame type (see Received Message Types)
	FrameId BYTE //1 +1  Frame ID (6 bit) + Parity (2 bit)
	Length BYTE //2 +2  Frame Length (1..8)
	Direction TLINDirection //3 +3  Frame Direction (see Message Direction Types)
	ChecksumType TLINChecksumType //4 +4  Frame Checksum type (see Message Checksum Types)
	Data [8]BYTE //5 +5  Data bytes (0..7)
	Checksum BYTE //6 +13 Frame Checksum
	ErrorFlags TLINMsgErrors //7 +16 Frame error flags (see Error flags for LIN Rcv Msgs)
	TimeStamp UINT64 //8 +24 Timestamp in microseconds
	hHw HLINHW //9 +32 Handle of the Hardware which received the message
}

// Represents a LIN Frame Entry
type TLINFrameEntry struct {
	// Size = 14 bytes
	FrameId BYTE //0 +0  Frame ID (without parity)
	Length BYTE //1 +1  Frame Length (1..8)
	Direction TLINDirection //2 +2  Frame Direction (see Message Direction Types)
	ChecksumType TLINChecksumType //3 +3  Frame Checksum type (see Message Checksum Types)
	Flags WORD //4 +4  Frame flags (see Frame flags for LIN Msgs)
	InitialData [8]BYTE //5 +6  Data bytes (0..7)

}

// Represents a LIN Schedule slot
type TLINScheduleSlot struct {
	// Size = 20 bytes
	Type TLINSlotType //0 +0  Slot Type (see Schedule Slot Types)
	Delay WORD //1 +2  Slot Delay in Milliseconds
	FrameId [8]BYTE //2 +4  Frame IDs (without parity)
	CountResolve BYTE //3 +12 ID count for sporadic frames, Resolve schedule number for Event frames
	Handle DWORD //4 +16 Slot handle (read-only)
}

// Represents LIN Status data
type TLINHardwareStatus struct {
	// Size = 8 bytes
	Mode TLINHardwareMode //0 +0  Hardware mode (see Hardware Operation Modes)
	Status TLINHardwareState //1 +1  Hardware status (see Hardware Status)
	FreeOnSendQueue	BYTE //2 +2  Count of free places in the Transmit Queue
	CountResolve WORD //3 +4  Free slots in the Schedule pool (see Minimum and Maximum values)
	Handle WORD
}