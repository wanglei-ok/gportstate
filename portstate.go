
// +build windows

package portstate

import (
	"unsafe"
	"syscall"
	"log"
	"fmt"
)

var (
	iphlpapi uintptr
	getTcpTable uintptr
	getTcpTable2 uintptr
	getUdpTable uintptr
)

const ANY_SIZE = 1
const (
	NO_ERROR = 0
	ERROR_INSUFFICIENT_BUFFER = 122
)


type DWORD uint32
type ULONG uint32

type MIB_TCP_STATE int32

const (
	MIB_TCP_STATE_CLOSED     MIB_TCP_STATE = 1
	MIB_TCP_STATE_LISTEN                   = 2
	MIB_TCP_STATE_SYN_SENT                 = 3
	MIB_TCP_STATE_SYN_RCVD                 = 4
	MIB_TCP_STATE_ESTAB                    = 5
	MIB_TCP_STATE_FIN_WAIT1                = 6
	MIB_TCP_STATE_FIN_WAIT2                = 7
	MIB_TCP_STATE_CLOSE_WAIT               = 8
	MIB_TCP_STATE_CLOSING                  = 9
	MIB_TCP_STATE_LAST_ACK                 = 10
	MIB_TCP_STATE_TIME_WAIT                = 11
	MIB_TCP_STATE_DELETE_TCB               = 12
)

type TCP_CONNECTION_OFFLOAD_STATE int32

const (
	TcpConnectionOffloadStateInHost TCP_CONNECTION_OFFLOAD_STATE = iota
	TcpConnectionOffloadStateOffloading
	TcpConnectionOffloadStateOffloaded
	TcpConnectionOffloadStateUploading
	TcpConnectionOffloadStateMax
)


type MIB_TCPTABLE struct {
	DwNumEntries DWORD
	Table        [ANY_SIZE]MIB_TCPROW
}

type PMIB_TCPTABLE *MIB_TCPTABLE

type MIB_TCPROW MIB_TCPROW_LH
type MIB_TCPROW_LH struct {
	State        MIB_TCP_STATE
	DwLocalAddr  DWORD
	DwLocalPort  DWORD
	DwRemoteAddr DWORD
	DwRemotePort DWORD
}

type MIB_TCPROW2 struct {
	DwState        DWORD
	DwLocalAddr    DWORD
	DwLocalPort    DWORD
	DwRemoteAddr   DWORD
	DwRemotePort   DWORD
	DwOwningPid    DWORD
	DwOffloadState TCP_CONNECTION_OFFLOAD_STATE
}

type MIB_TCPTABLE2 struct {
	DwNumEntries DWORD
	Table        [ANY_SIZE]MIB_TCPROW2
}

type PMIB_TCPTABLE2 *MIB_TCPTABLE2


type MIB_UDPTABLE struct {
	DwNumEntries DWORD
	Table        [ANY_SIZE]MIB_UDPROW
}

type PMIB_UDPTABLE *MIB_UDPTABLE

type MIB_UDPROW struct {
	DwLocalAddr DWORD
	DwLocalPort DWORD
}

func init(){
	iphlpapi = loadLibrary("iphlpapi.dll")

	getTcpTable = getProcAddress(iphlpapi,"GetTcpTable")
	getTcpTable2 = getProcAddress(iphlpapi,"GetTcpTable2")
	getUdpTable = getProcAddress(iphlpapi,"GetUdpTable")
}

func loadLibrary(libname string) uintptr {
	handle, err := syscall.LoadLibrary(libname)
	if err != nil {
		log.Fatalf("LoadLibrary %s error:%v",libname, err)
	}
	return uintptr(handle)
}

func getProcAddress(lib uintptr, name string) uintptr {
	addr, err := syscall.GetProcAddress(syscall.Handle(lib), name)
	if err != nil {
		log.Fatalf("GetProcAddress %s error:%v",name, err)
	}
	return uintptr(addr)
}


func getUintptrFromBool(b bool) uintptr {
	if b {
		return 1
	} else {
		return 0
	}
}

func GetTcpTable(tcpTable PMIB_TCPTABLE, sizePointer *DWORD, order bool) DWORD {
	ret1,_,_ := syscall.Syscall(getTcpTable, 3,
		uintptr(unsafe.Pointer(tcpTable)),
		uintptr(unsafe.Pointer(sizePointer)),
		getUintptrFromBool(order))
	return DWORD(ret1)
}

func GetTcpTable2(tcpTable PMIB_TCPTABLE2, sizePointer *DWORD, order bool) DWORD {
	ret1,_,_ := syscall.Syscall(getTcpTable2, 3,
		uintptr(unsafe.Pointer(tcpTable)),
		uintptr(unsafe.Pointer(sizePointer)),
		getUintptrFromBool(order))
	return DWORD(ret1)
}

func GetUdpTable(udpTable PMIB_UDPTABLE, sizePointer *DWORD, order bool) ULONG {
	ret1,_,_ := syscall.Syscall(getUdpTable, 3,
		uintptr(unsafe.Pointer(udpTable)),
		uintptr(unsafe.Pointer(sizePointer)),
		getUintptrFromBool(order))
	return ULONG(ret1)
}


func GetTcpPortState( nPort ULONG) (MIB_TCP_STATE, bool){
	var tcpTable[100] MIB_TCPTABLE
	var nSize DWORD = DWORD(unsafe.Sizeof(tcpTable))

	if NO_ERROR == GetTcpTable( &tcpTable[0], &nSize,true) {
		nCount := tcpTable[0].DwNumEntries
		if nCount > 0	{
			ptr := uintptr(unsafe.Pointer(&tcpTable[0].Table[0]))
			for i := DWORD(0);i<nCount;i++ {
				tcpRow := *(*MIB_TCPROW)(unsafe.Pointer(ptr))
				temp1 := tcpRow.DwLocalPort
				temp2 := temp1 / 256 + (temp1 % 256) * 256
				if temp2 == DWORD(nPort) {
					return tcpRow.State, true
				}

				ptr+=unsafe.Sizeof(tcpRow)
			}
		}
		return 0,false
	}
	return 0,false
}

func GetUdpPortState( nPort ULONG ) bool {
	var udpTable[100] MIB_UDPTABLE
	var nSize = DWORD(unsafe.Sizeof(udpTable))

	if NO_ERROR == GetUdpTable(&udpTable[0],&nSize,true) {
		nCount := udpTable[0].DwNumEntries
		if  nCount > 0 {
			ptr := uintptr(unsafe.Pointer(&udpTable[0].Table[0]))
			for i:=DWORD(0);i<nCount;i++ {
				ucpRow := *(*MIB_UDPROW)(unsafe.Pointer(ptr))
				temp1 := ucpRow.DwLocalPort
				temp2 := temp1 / 256 + (temp1 % 256) * 256
				if temp2 == DWORD(nPort) {
					return true
				}

				ptr += unsafe.Sizeof(ucpRow)
			}
		}
		return false
	}
	return false
}

var tcpStateStrings = []string{
"???",
"CLOSED",
"LISTENING",
"SYN_SENT",
"SEN_RECEIVED",
"ESTABLISHED",
"FIN_WAIT",
"FIN_WAIT2",
"CLOSE_WAIT",
"CLOSING",
"LAST_ACK",
"TIME_WAIT",
}


func EnumTCPTable() DWORD {
	var pTcpTable *MIB_TCPTABLE = nil
	var dwSize DWORD = 0
	var dwRetVal DWORD = NO_ERROR
	var buf *memBuffer

	if GetTcpTable(pTcpTable, &dwSize, true) == ERROR_INSUFFICIENT_BUFFER {
		buf = alloc( uint32(dwSize) )
		pTcpTable = PMIB_TCPTABLE(buf.ptr)
	} else {
		return dwRetVal
	}

	fmt.Printf("Active Connections\n\n")
	fmt.Printf("  Proto\t%-24s%-24s%s\n","Local Address","Foreign Address","State")

	if dwRetVal = GetTcpTable(pTcpTable, &dwSize, true); dwRetVal == NO_ERROR {
		ptr := uintptr(unsafe.Pointer(&pTcpTable.Table[0]))
		for i := DWORD(0); i < pTcpTable.DwNumEntries; i++ {
			tcpRow := *(*MIB_TCPROW)(unsafe.Pointer(ptr))

			if tcpRow.State == MIB_TCP_STATE_LISTEN {
				tcpRow.DwRemotePort = 0
			}

			strlip := fmt.Sprintf("%s:%d",inet_ntoa(uint32(tcpRow.DwLocalAddr)),ntohs(uint16(tcpRow.DwLocalPort)))
			strrip := fmt.Sprintf("%s:%d",inet_ntoa(uint32(tcpRow.DwRemoteAddr)),ntohs(uint16(tcpRow.DwRemotePort)))
			fmt.Printf("  TCP\t%-24s%-24s%s\n",strlip,strrip,tcpStateStrings[tcpRow.State])

			ptr += unsafe.Sizeof(tcpRow)
		}
	} else {
		fmt.Printf("\tCall to GetTcpTable failed.\n")

		var flags uint32 = syscall.FORMAT_MESSAGE_ALLOCATE_BUFFER | syscall.FORMAT_MESSAGE_IGNORE_INSERTS | syscall.FORMAT_MESSAGE_FROM_SYSTEM
		libpdhDll := syscall.MustLoadDLL("pdh.dll")
		buffer := make([]uint16, 300)

		// Have to cast Handle, which is a uintptr, to a uint32 due to the signature of FormatMessage
		_, err := syscall.FormatMessage(flags, uint32(libpdhDll.Handle), uint32(dwRetVal), 0, buffer, nil)
		if err != nil {

		}else{
			fmt.Printf("\tError: %s", buffer)
		}
	}

	buf.free()
	return dwRetVal
}

func EnumTCPTable2() DWORD {
	var pTcpTable *MIB_TCPTABLE2 = nil
	var dwSize DWORD = 0
	var dwRetVal DWORD = NO_ERROR
	var buf *memBuffer

	snapshot := NewSnapshot()
	snapshot.Print()

	//获得pTcpTable所需要的真实长度,dwSize
	if GetTcpTable2(pTcpTable, &dwSize, true) == ERROR_INSUFFICIENT_BUFFER {
		buf = alloc( uint32(dwSize) )
		pTcpTable = PMIB_TCPTABLE2(buf.ptr)
	} else {
		return dwRetVal
	}

	fmt.Printf("Active Connections\n\n")
	fmt.Printf("  Proto\t%-24s%-24s%-15s%s\n","Local Address","Foreign Address","State","Pid")

	if dwRetVal = GetTcpTable2(pTcpTable, &dwSize, true); dwRetVal == NO_ERROR {
		ptr := uintptr(unsafe.Pointer(&pTcpTable.Table[0]))
		for i := DWORD(0); i < pTcpTable.DwNumEntries; i++ {
			tcpRow := *(*MIB_TCPROW2)(unsafe.Pointer(ptr))
			if tcpRow.DwState == MIB_TCP_STATE_LISTEN {
				tcpRow.DwRemotePort = 0
			}

			strlip := fmt.Sprintf("%s:%d",inet_ntoa(uint32(tcpRow.DwLocalAddr)),ntohs(uint16(tcpRow.DwLocalPort)))
			strrip := fmt.Sprintf("%s:%d",inet_ntoa(uint32(tcpRow.DwRemoteAddr)),ntohs(uint16(tcpRow.DwRemotePort)))
			fmt.Printf("  TCP\t%-24s%-24s%-15s%s\n",strlip,strrip,tcpStateStrings[tcpRow.DwState],snapshot.Name(ulong(tcpRow.DwOwningPid)))

			ptr += unsafe.Sizeof(tcpRow)
		}
	} else {
		fmt.Printf("\tCall to GetTcpTable failed.\n")

		var flags uint32 = syscall.FORMAT_MESSAGE_ALLOCATE_BUFFER | syscall.FORMAT_MESSAGE_IGNORE_INSERTS | syscall.FORMAT_MESSAGE_FROM_SYSTEM
		libpdhDll := syscall.MustLoadDLL("pdh.dll")
		buffer := make([]uint16, 300)

		// Have to cast Handle, which is a uintptr, to a uint32 due to the signature of FormatMessage
		_, err := syscall.FormatMessage(flags, uint32(libpdhDll.Handle), uint32(dwRetVal), 0, buffer, nil)
		if err != nil {

		}else{
			fmt.Printf("\tError: %s", buffer)
		}
	}

	buf.free()
	return dwRetVal
}


func EnumTCPTable2ForPort( port DWORD ) DWORD {
	var pTcpTable *MIB_TCPTABLE2 = nil
	var dwSize DWORD = 0
	var dwRetVal DWORD = NO_ERROR
	var buf *memBuffer

	port = DWORD(ntohs(uint16(port)))

	snapshot := NewSnapshot()


	//获得pTcpTable所需要的真实长度,dwSize
	if GetTcpTable2(pTcpTable, &dwSize, true) == ERROR_INSUFFICIENT_BUFFER {
		buf = alloc( uint32(dwSize) )
		pTcpTable = PMIB_TCPTABLE2(buf.ptr)
	} else {
		return dwRetVal
	}

	fmt.Printf("Active Connections\n\n")
	fmt.Printf("  Proto\t%-24s%-24s%-15s%-10s%s\n","Local Address","Foreign Address","State","Pid", "Path")

	if dwRetVal = GetTcpTable2(pTcpTable, &dwSize, true); dwRetVal == NO_ERROR {
		ptr := uintptr(unsafe.Pointer(&pTcpTable.Table[0]))
		for i := DWORD(0); i < pTcpTable.DwNumEntries; i++ {
			tcpRow := *(*MIB_TCPROW2)(unsafe.Pointer(ptr))
			ptr += unsafe.Sizeof(tcpRow)

			if tcpRow.DwState != MIB_TCP_STATE_ESTAB {
				continue
			}

			if tcpRow.DwLocalPort != port {
				continue
			}

			if tcpRow.DwState == MIB_TCP_STATE_LISTEN {
				tcpRow.DwRemotePort = 0
			}

			strlip := fmt.Sprintf("%s:%d",inet_ntoa(uint32(tcpRow.DwLocalAddr)),ntohs(uint16(tcpRow.DwLocalPort)))
			strrip := fmt.Sprintf("%s:%d",inet_ntoa(uint32(tcpRow.DwRemoteAddr)),ntohs(uint16(tcpRow.DwRemotePort)))
			fmt.Printf("  TCP\t%-24s%-24s%-15s%-10d%s\n",strlip,strrip,tcpStateStrings[tcpRow.DwState],tcpRow.DwOwningPid, snapshot.Name(ulong(tcpRow.DwOwningPid)))

		}
	} else {
		fmt.Printf("\tCall to GetTcpTable failed.\n")

		var flags uint32 = syscall.FORMAT_MESSAGE_ALLOCATE_BUFFER | syscall.FORMAT_MESSAGE_IGNORE_INSERTS | syscall.FORMAT_MESSAGE_FROM_SYSTEM
		libpdhDll := syscall.MustLoadDLL("pdh.dll")
		buffer := make([]uint16, 300)

		// Have to cast Handle, which is a uintptr, to a uint32 due to the signature of FormatMessage
		_, err := syscall.FormatMessage(flags, uint32(libpdhDll.Handle), uint32(dwRetVal), 0, buffer, nil)
		if err != nil {

		}else{
			fmt.Printf("\tError: %s", buffer)
		}
	}

	buf.free()

	snapshot.Print()
	return dwRetVal
}


