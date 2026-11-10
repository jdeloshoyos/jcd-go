'
'  pb_proc.bas
'
'  functions for finding running processes
'  public domain from Don Dickinson
'
#if not %def(%PB_PROC_BAS)
%PB_PROC_BAS = 1

%TH32CS_SNAPHEAPLIST = 1
%TH32CS_SNAPPROCESS = 2
%TH32CS_SNAPTHREAD = 4
%TH32CS_SNAPMODULE = 8
%TH32CS_SNAPALL = 15
%TH32CS_INHERIT = &h80000000

type PROCESSENTRY32
	dwSize as dword
	cntUsage as dword
	th32ProcessID as dword
	th32DefaultHeapID as dword ptr
	th32ModuleID as dword
	cntThreads as dword
	th32ParentProcessID as dword
	pcPriClassBase as long
	dwFlags as long
	szExeFile as asciiz * %MAX_PATH
end type

declare function CreateToolhelp32Snapshot lib "kernel32.dll" alias "CreateToolhelp32Snapshot" _
( ByVal dwFlags as dword, ByVal th32processID as dword ) as long
declare function Process32First lib "kernel32.dll" alias "Process32First" _
( ByVal hHandle as long, rProc as PROCESSENTRY32 ) as long
declare function Process32Next lib "kernel32.dll" alias "Process32Next" _
( ByVal hHandle as long, rProc as PROCESSENTRY32 ) as long


'~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
' procGetList
'
' fills the rProc() array with the processentry32 information for
' all running processes.
' returns %true on success, %false on failure
'~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
function procGetList alias "procGetList" _
( rProc() as PROCESSENTRY32 ) export as long
	
	dim iResult as long
	dim iCount as long
	dim hHandle as long
	dim hProc as long
	dim oneProc as PROCESSENTRY32
	
	redim rProc(0 to 0) as PROCESSENTRY32

	'- default to fail.
	function = 0
	iCount = 0

	'- Get a snapshot of what's running
	hHandle = CreateToolhelp32Snapshot(%TH32CS_SNAPALL, 0)
	if hHandle = -1 then
	? "CreateToolHelp32Snapshot failed"
	exit function
	end if

	'- loop through everything
	oneProc.dwSize = sizeof(oneProc)
	iResult = Process32First(hHandle, oneProc)
	Do until iResult = 0
	incr iCount
	redim preserve rProc(0 to iCount) as PROCESSENTRY32
	rProc(iCount) = oneProc
	iResult = Process32Next(hHandle, oneProc)
	Loop
	
	closehandle hHandle
	function = %true

end function

'~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
' procFind
'
' returns the process id of the module sent.
' returns -1 if it can't be located.
'~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
function procGetID alias "procGetID" ( zModule as asciiz ) export as long
	
	dim iResult as long
	dim hHandle as long
	dim hProc as long
	dim rProc as PROCESSENTRY32

	'- default to fail.
	function = -1
	
	zModule = trim$(lcase$(zModule))

	'- Get a snapshot of what's running
	hHandle = CreateToolhelp32Snapshot(%TH32CS_SNAPALL, 0)
	
	if hHandle <> -1 then

	'- Try to locate the module
	rProc.dwSize = sizeof(rProc)
	iResult = Process32First(hHandle, rProc)
	Do until iResult = 0
	if trim$(lcase$(rProc.szExeFile)) = zModule then
	function = rProc.th32ProcessID
	exit do
	end if
	iResult = Process32Next(hHandle, rProc)
	Loop
	
	closehandle hHandle
	end if

end function

'~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
' procKill
'
' pass it a process ID and it kills that process
' returns %true if the process could be opened.
' returns %false if it could not.
'~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
function procKill alias "procKill" ( byval procID as long ) export as long
	
	dim hProc as long
	
	function = %false
	
	hProc = OpenProcess(%PROCESS_TERMINATE, %false, procID)
	if hProc then
	terminateProcess hProc, 0
	closehandle hProc
	function = %true
	end if

end function

#endif