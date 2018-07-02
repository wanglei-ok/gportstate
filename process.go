package portstate

import (
	"unsafe"
	"fmt"
	"strconv"
	"syscall"
)

type ulong int32
type ulong_ptr uintptr

type PROCESSENTRY32 struct {
	dwSize ulong
	cntUsage ulong
	th32ProcessID ulong
	th32DefaultHeapID ulong_ptr
	th32ModuleID ulong
	cntThreads ulong
	th32ParentProcessID ulong
	pcPriClassBase ulong
	dwFlags ulong
	szExeFile [260]byte
}

type MODULEENTRY32 struct {
	dwSize         ulong
	moduleID     ulong
	processID    ulong
	glblcntUsage ulong
	proccntUsage ulong
	modBaseAddr  *byte
	modBaseSize  ulong
	hModule      uintptr
	//szUnknown    [16]byte
	szModule     [255+1]byte
	szExePath    [260]byte
}

type ProcessInfo struct {
	Name string
	Path string
}

type Snapshot map[ulong] ProcessInfo

func NewSnapshot() *Snapshot{
	ret := make(Snapshot,0)

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	CreateToolhelp32Snapshot := kernel32.NewProc("CreateToolhelp32Snapshot")
	pHandle,_,_ := CreateToolhelp32Snapshot.Call(uintptr(0x2),uintptr(0x0))
	if int(pHandle)==-1 {
		return &ret
	}
	Process32Next := kernel32.NewProc("Process32Next")
	for {
		var proc PROCESSENTRY32
		proc.dwSize = ulong(unsafe.Sizeof(proc))
		if rt,_,_ := Process32Next.Call(uintptr(pHandle),uintptr(unsafe.Pointer(&proc)));int(rt)==1 {
			var pi ProcessInfo
			pi.Name = string(proc.szExeFile[0:])
			pi.Path = getProcessPath(proc.th32ProcessID)
			ret[proc.th32ProcessID] = pi
		}else{
			break
		}
	}
	CloseHandle := kernel32.NewProc("CloseHandle")
	_,_,_ = CloseHandle.Call(pHandle)

	return &ret
}

func (s *Snapshot) Name(pid ulong) string {
	pi, ok := (*s)[ulong(pid)]
	if ok {
		return fmt.Sprintf("%s", pi.Name)
	}
	return fmt.Sprintf("[%d]", pid)
}

func (s *Snapshot) Print() {
	fmt.Println("ProcessList")
	for pid, info := range *s {
		fmt.Println("ProcessID : "+strconv.Itoa(int(pid)))
		fmt.Println("ProcessName : "+info.Name)
		fmt.Println("ProcessPath : "+info.Path)
		fmt.Println("")
	}
}


func PrintProcessInfo() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	CreateToolhelp32Snapshot := kernel32.NewProc("CreateToolhelp32Snapshot")
	pHandle,_,_ := CreateToolhelp32Snapshot.Call(uintptr(0x2),uintptr(0x0))
	if int(pHandle)==-1 {
		return
	}
	Process32Next := kernel32.NewProc("Process32Next")
	for {
		var proc PROCESSENTRY32
		proc.dwSize = ulong(unsafe.Sizeof(proc))
		if rt,_,_ := Process32Next.Call(uintptr(pHandle),uintptr(unsafe.Pointer(&proc)));int(rt)==1 {
			fmt.Println("ProcessName : "+string(proc.szExeFile[0:]))
			fmt.Println("ProcessID : "+strconv.Itoa(int(proc.th32ProcessID)))
		}else{
			break
		}
	}
	CloseHandle := kernel32.NewProc("CloseHandle")
	_,_,_ = CloseHandle.Call(pHandle)
}

func getProcessPath(pid ulong) string {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	CreateToolhelp32Snapshot := kernel32.NewProc("CreateToolhelp32Snapshot")
	pHandle,_,_ := CreateToolhelp32Snapshot.Call(uintptr(0x8),uintptr(pid))
	if int(pHandle)==-1 {
		return ""
	}
	Module32First := kernel32.NewProc("Module32First")
	var mod MODULEENTRY32
	mod.dwSize = ulong(unsafe.Sizeof(mod))
	if rt,_,_ := Module32First.Call(uintptr(pHandle),uintptr(unsafe.Pointer(&mod)));int(rt)==1 {
		fmt.Printf("szModule"+string(mod.szModule[0:]))
		return string(mod.szExePath[0:]) //stringFromUnicode16(&mod.szExePath[0])
	}
	CloseHandle := kernel32.NewProc("CloseHandle")
	_,_,_ = CloseHandle.Call(pHandle)
	return ""
}


