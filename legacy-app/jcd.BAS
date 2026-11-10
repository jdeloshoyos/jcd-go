'SED_PBCC

' Visual dir navigator
' Jaime de los Hoyos, 2010

' LICENSE TERMS - DO NOT REMOVE FROM SOURCE!
'
'Permission is hereby granted, free of charge, to any person obtaining a copy
'of this software and associated documentation files (the "Software"), to deal
'in the Software without restriction, including without limitation the rights
'to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
'copies of the Software, and to permit persons to whom the Software is
'furnished to do so, subject to the following conditions:
'
'The above copyright notice and this permission notice shall be included in
'all copies or substantial portions of the Software.
'
'THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
'IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
'FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
'AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
'LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
'OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
'THE SOFTWARE.

#compile exe "jcd_bin.exe"
#dim all

$version="1.2"
$drive_sel_label="System drives"
%max_search_length=30

' Version history:
' 1.07 : New method for changing directory on exit, using a batch file and without the ugly KillParent hack
' 1.08 : Minor fixes
' 1.09 : Changed the output and the batch file for better error handling
' 1.10 : Escape certain special characters, necessary for the FOR /F command in batch files
' 1.11 : Hidden directories are listed now
' 1.12 : Check that directory has at least one entry before accepting ENTER key
' 1.13 : Remove trailing slashes from path passed as argument
' 1.14 : Handle path passed as argument if it ends in a period
' 1.15 : Search pattern shown in footer
' 1.16 : Check for calling shell as first argument, and replace certain characters in output path for cmd.exe
' 1.17 : Fixed bug; pressing modifier keys (accents, etc.) froze the program with next keypress
' 1.18 : Included "l" letter for symlinks/reparse points in file attributes
' 1.2  : Info about entry attributes in help. Search now searches for matches in any part of the entry

#include "win32api.inc"

' Functions for working with processes, etc.
' Public domain from Don Dickinson, thanks a lot!!!
#include "pb_proc.inc"

function pbmain
	
	local start_dir as string
	local oDir as iDirectoryInfo
	local oCons as iConsole
	local i as long
	local current_position as long
	local current_offset as long
	local keypressed as string
	local buffer as string
	local newshell_pid as long
	local search_str as string
	local last_search_time as long
	local search_index as long
	local sort_mode as long
	local calling_shell as string
	
	' Initialize the DirEntries and Cons objects
	let oDir=class "DirectoryInfo"
	let oCons=class "ConsoleClass"
	
	search_index=1
	sort_mode=1
	
	' v1.16: Check argument that specifies calling shell (c=cmd.exe, p=powershell)
	calling_shell=ucase$(command$(1))
	
	if calling_shell<>"C" and calling_shell<>"P" then
		' Invalid calling shell argument
		stdout $version+"*ERROR*Invalid calling shell argument, should be C or P"
		exit function
	end if
	
	' Set starting dir. If none was specified, we use current dir.
	start_dir=trim$(command$(2), $DQ)
	
	' Some checks if a command line parameter was passed
	if start_dir<>"" then
		
		' Did the user specify a dir starting with "\"? That means a dir on the current drive's root.
		if left$(start_dir, 1)="\" then start_dir=left$(curdir$, 2)+start_dir
		
		' If the specified dir was not a full path, we convert it to that.
		if mid$(start_dir, 2, 1)<>":" then start_dir=curdir$+"\"+start_dir
		
		' v1.13: Trim trailing slashes/backslashes
		start_dir = rtrim$(start_dir, any "\/")
		
	else
		start_dir=curdir$
	end if
	
	if oCons.Width>=80 and oCons.Height>=20 then
		' To avoid unpleasant results, this program requires a console width of 80 and height of 20 at minimum.
		if isfolder(start_dir) then
			' The starting dir is valid, we are good to go
			chdir start_dir
			start_dir=curdir$	' v1.14: Handle "relative" paths passed as arguments (ending in . or .., specifically)
			
			' Initialize position/offset variables
			current_position=1
			current_offset=0
			
			' Let's get the current directory info
			oDir.ScanDir(start_dir)
			
			' Prepare the screen and draw the header
			cls
			oCons.DrawHeader(start_dir, oDir)
			oCons.DrawOffset(current_position, oDir.NumEntries, sort_mode)
			oCons.DrawFooter(search_str)
			oCons.DrawScreen(current_offset, oDir)
			oCons.DrawEntry(current_position, current_position-current_offset+2, 1, oDir)	' Highlight the first entry
			
			' Main program loop. Each iteration checks for key presses and updates the screen accordingly.
			do
				' Inner loop to catch keypresses
				keypressed=waitkey$
				select case keypressed
					case is chr$(27)
						exit do	' ESC was pressed
					
					case chr$(0)+chr$(72)	' Arrow up
						if current_position>1 then
							if current_position>current_offset+1 then
								' Do not scroll, simply update the highlighted item
								oCons.DrawEntry(current_position, current_position-current_offset+2, 0, oDir)	' De-Highlight the current entry
								decr current_position
								oCons.DrawEntry(current_position, current_position-current_offset+2, 1, oDir)	' Highlight the new entry
								oCons.DrawOffset(current_position, oDir.NumEntries, sort_mode)
							else
								' We have to scroll
								decr current_offset
								decr current_position
								oCons.DrawScreen(current_offset, oDir)
								oCons.DrawEntry(current_position, current_position-current_offset+2, 1, oDir)	' Highlight the first entry
								oCons.DrawOffset(current_position, oDir.NumEntries, sort_mode)
							end if
						end if
					
					case chr$(0)+chr$(80)	' Arrow down
						if current_position<oDir.NumEntries then
							if current_position<current_offset+oCons.Height-3 then
								' Do not scroll, simply update the highlighted item
								oCons.DrawEntry(current_position, current_position-current_offset+2, 0, oDir)	' De-Highlight the current entry
								incr current_position
								oCons.DrawEntry(current_position, current_position-current_offset+2, 1, oDir)	' Highlight the new entry
								oCons.DrawOffset(current_position, oDir.NumEntries, sort_mode)
							else
								' We have to scroll
								incr current_offset
								incr current_position
								oCons.DrawScreen(current_offset, oDir)
								oCons.DrawEntry(current_position, current_position-current_offset+2, 1, oDir)	' Highlight the first entry
								oCons.DrawOffset(current_position, oDir.NumEntries, sort_mode)
							end if
						end if
					
					case chr$(8)	' Backspace
						' Clear search parameters
						search_str=""
						search_index=1
						
						' First we check if we are already at a drive's root.
						regexpr "^[A-Z]:\\$" in start_dir to i
						if i=1 or start_dir=$drive_sel_label then
							' The previous regex got a match for a drive's root.
							start_dir=$drive_sel_label
							current_position=1
							current_offset=0
							oDir.ScanDrives
							cls
							oCons.DrawHeader(start_dir, oDir)
							oCons.DrawOffset(current_position, oDir.NumEntries, sort_mode)
							oCons.DrawFooter(search_str)
							oCons.DrawScreen(current_offset, oDir)
							oCons.DrawEntry(current_position, current_position-current_offset+2, 1, oDir)	' Highlight the first entry
						else
							chdir ".."
							start_dir=curdir$
							current_position=1
							current_offset=0
							oDir.ScanDir(start_dir)
							cls
							oCons.DrawHeader(start_dir, oDir)
							oCons.DrawOffset(current_position, oDir.NumEntries, sort_mode)
							oCons.DrawFooter(search_str)
							oCons.DrawScreen(current_offset, oDir)
							oCons.DrawEntry(current_position, current_position-current_offset+2, 1, oDir)	' Highlight the first entry
						end if
					
					case chr$(13)	' Enter
						' v1.12: First we check that the current dir actually has at least one entry, or else we'll get *very* undefined results
						if oDir.NumEntries > 0 then
							' Clear search parameters
							search_str=""
							search_index=1
							
							if start_dir=$drive_sel_label then
								' We were at the drive selection screen
								chdrive oDir.GetEntryName(current_position)
								chdir "\"
								start_dir=curdir$
								current_position=1
								current_offset=0
								oDir.ScanDir(start_dir)
								cls
								oCons.DrawHeader(start_dir, oDir)
								oCons.DrawOffset(current_position, oDir.NumEntries, sort_mode)
								oCons.DrawFooter(search_str)
								oCons.DrawScreen(current_offset, oDir)
								oCons.DrawEntry(current_position, current_position-current_offset+2, 1, oDir)	' Highlight the first entry
							elseif left$(oDir.GetEntryAttr(current_position), 1)="d" then
								' It is a directory
								chdir oDir.GetEntryName(current_position)
								start_dir=curdir$
								current_position=1
								current_offset=0
								oDir.ScanDir(start_dir)
								cls
								oCons.DrawHeader(start_dir, oDir)
								oCons.DrawOffset(current_position, oDir.NumEntries, sort_mode)
								oCons.DrawFooter(search_str)
								oCons.DrawScreen(current_offset, oDir)
								oCons.DrawEntry(current_position, current_position-current_offset+2, 1, oDir)	' Highlight the first entry
							else
								' It's a file, we try to open it as if from Windows Explorer
								ShellExecute (%NULL, "open", start_dir+"\"+oDir.GetEntryName(current_position), "", "", %SW_SHOWNORMAL)
							end if
						end if
					
					case chr$(0)+chr$(73)	' PageUp
						if oDir.NumEntries>oCons.Height-3 then
							current_offset=max(current_offset-(oCons.Height-3), 0)
							current_position=current_offset+1
							oCons.DrawScreen(current_offset, oDir)
							oCons.DrawEntry(current_position, current_position-current_offset+2, 1, oDir)	' Highlight the first entry
							oCons.DrawOffset(current_position, oDir.NumEntries, sort_mode)
						end if
					
					case chr$(0)+chr$(81)	' PageDown
						if oDir.NumEntries>oCons.Height-3 then
							current_offset=min(current_offset+(oCons.Height-3), oDir.NumEntries-(oCons.Height-3))
							current_position=current_offset+1
							oCons.DrawScreen(current_offset, oDir)
							oCons.DrawEntry(current_position, current_position-current_offset+2, 1, oDir)	' Highlight the first entry
							oCons.DrawOffset(current_position, oDir.NumEntries, sort_mode)
						end if
					
					case chr$(0)+chr$(59)	' F1
						' Show help.
						oCons.ShowHelp
						
					case chr$(0)+chr$(60)	' F2
						' Toggle sort mode. Available: Name, Date of last write, Size, in ascending and descending order.
						incr sort_mode
						if sort_mode>6 then sort_mode=1
						oDir.SortMode=sort_mode
						
						if start_dir=$drive_sel_label then
							' We only do an offset label update.
							oCons.DrawOffset(current_position, oDir.NumEntries, sort_mode)
						else
							
							' We do a full refresh, re-reading the directory entries and sorting them accordingly.
							
							' Clear search parameters
							search_str=""
							search_index=1
							
							start_dir=curdir$
							current_position=1
							current_offset=0
							oDir.ScanDir(start_dir)
							cls
							oCons.DrawHeader(start_dir, oDir)
							oCons.DrawOffset(current_position, oDir.NumEntries, sort_mode)
							oCons.DrawFooter(search_str)
							oCons.DrawScreen(current_offset, oDir)
							oCons.DrawEntry(current_position, current_position-current_offset+2, 1, oDir)	' Highlight the first entry
						
						end if
						
					case chr$(32) to chr$(126)		' Any character within the printable range.
						i=GetTickCount		' Check milliseconds elapsed since last boot.
						if i-last_search_time<1000 then
							if len(search_str) < %max_search_length then
								' We add the character to the last search performed
								last_search_time=i
								search_str=search_str+keypressed
								search_index=1
								oCons.DrawFooter(search_str)
							else
								beep
							end if
						else
							' Too much time passed since the last entered search, we start a new one instead
							last_search_time=i
							search_str=keypressed
							search_index=1
							oCons.DrawFooter(search_str)
						end if
						
						' We search within the current dir entries if there's a match starting with search_dir.
						' If there is one, return the index number of the FIRST match and select it.
						if oDir.SearchEntry(search_str, 0)>0 then		' Get matches count
							i=oDir.SearchEntry(search_str, search_index)		' Get (search_index)-th match
							
							if i>current_offset and i<=current_offset+oCons.Height-3 then
								' The entry is already on screen, we just highlight it.
								oCons.DrawEntry(current_position, current_position-current_offset+2, 0, oDir)	' De-Highlight the current entry
								current_position=i
								oCons.DrawEntry(current_position, current_position-current_offset+2, 1, oDir)	' Highlight the new entry
								oCons.DrawOffset(current_position, oDir.NumEntries, sort_mode)
							else
								' The entry is offscreen, we have to scroll just enough to make it visible.
								if i<=current_offset then
									' We have to scroll "up"
									current_offset=i-1
								else
									' We have to scroll "down"
									current_offset=i-oCons.Height+3
								end if
								current_position=i
								oCons.DrawScreen(current_offset, oDir)
								oCons.DrawEntry(current_position, current_position-current_offset+2, 1, oDir)	' Highlight the first entry
								oCons.DrawOffset(current_position, oDir.NumEntries, sort_mode)
							end if
							
						else
							' Match not found!
							beep
						end if
						
					case $tab	' Tab was pressed to cycle between search results
						if oDir.SearchEntry(search_str, 0)>0 then		' Get matches count
							incr search_index
							if search_index>oDir.SearchEntry(search_str, 0) then search_index=1	' Loop through results
							
							i=oDir.SearchEntry(search_str, search_index)		' Get (search_index)-th match
							
							if i>current_offset and i<=current_offset+oCons.Height-3 then
								' The entry is already on screen, we just highlight it.
								oCons.DrawEntry(current_position, current_position-current_offset+2, 0, oDir)	' De-Highlight the current entry
								current_position=i
								oCons.DrawEntry(current_position, current_position-current_offset+2, 1, oDir)	' Highlight the new entry
								oCons.DrawOffset(current_position, oDir.NumEntries, sort_mode)
							else
								' The entry is offscreen, we have to scroll just enough to make it visible.
								if i<=current_offset then
									' We have to scroll "up"
									current_offset=i-1
								else
									' We have to scroll "down"
									current_offset=i-oCons.Height+3
								end if
								current_position=i
								oCons.DrawScreen(current_offset, oDir)
								oCons.DrawEntry(current_position, current_position-current_offset+2, 1, oDir)	' Highlight the first entry
								oCons.DrawOffset(current_position, oDir.NumEntries, sort_mode)
							end if
							
						else
							' No matches
							beep
						end if
						
					case else
						input flush  ' 1.17 - Fix bug with modifier keys (accents)
						' For debugging only, print the dec codes of the keypress.
						' REMEMBER TO COMMENT OUT!
						'for i=1 to len(keypressed)
						'	stdout format$(asc(keypressed, i))+" ";
						'next i
						'stdout
					
				end select
				
			loop
			
			' Clean up the screen and go to last directory specified
			color 7,0
			cls
			'oCons.PrintCredits
			
			' Create  new cmd.exe instance starting in the specified dir
			' If we were last looking at the drive selection screen, set start_dir to the last current directory instead.
			if start_dir=$drive_sel_label then start_dir=curdir$
			'newshell_pid=shell(environ$("COMSPEC")+" /K "+$DQ+"cd "+start_dir+"&&title %COMSPEC%"+$DQ)
			'KillParent	' This sounds harsher than it really is...
			
			' v1.10: We need to escape certain characters that can break the FOR /F in the batch file.
			' See https://www.robvanderwoude.com/escapechars.php for reference
			' v1.16: This only applies if the calling shell is cmd.exe, ignore for powershell
			if calling_shell="C" then
				replace "'" with "^'" in start_dir
				replace "," with "^," in start_dir
				replace ";" with "^;" in start_dir
				replace "=" with "^=" in start_dir
				replace "(" with "^(" in start_dir
				replace ")" with "^)" in start_dir
			end if
			
			stdout $version+"*OK*"+start_dir	' v1.07 - This outputs the directory to change to. The batch file takes care of the rest.
			
		else
			' The directory passed as a command line parameter does not exist
			stdout $version+"*ERROR*"+start_dir+" is not a valid directory or does not exist."
		end if
	else
		' The minimum console dimensions are not met (!!! seriously, what kind of console do you have??)
		' I mean, less than 80 x 20!? Come on!
		stdout $version+"*ERROR*A console size of at least 80 columns and 20 rows is required."
	end if
	
	' Dereference the created objects
	let oDir=nothing
	let oCons=nothing
	
end function


' ---------------------------------------
'  Procedures and classes
' ---------------------------------------

#include "DirectoryInfo_class.inc"
#include "Console_class.inc"

function HumanFileSize(byval bytes as quad) as string
	' Converts a file size in bytes to human-readable form.
	' eg. 1024 --> 1 kb; 2000000 --> 2 mb (a bit less, actually), etc.
	
	local level as long
	local result as string
	local bytesfloat as ext
	
	bytesfloat=bytes        ' We operate on this ext to get fractional numbers.
	
	do while bytesfloat>1024
		incr level
		bytesfloat=bytesfloat/1024
	loop
	
	' The level variable tells us how many divisions we've made. Based on that, we can determine the order of magnitude (bytes, kb, mb, etc.)
	' Up to yottabyte should suffice :)
	result=switch$(level=0, " bytes", _
						level=1, " KB", _
						level=2, " MB", _
						level=3, " GB", _
						level=4, " TB", _
						level=5, " PB", _
						level=6, " EB", _
						level=7, " ZB", _
						level=8, " YB")
	
	if level=0 then
		result=ltrim$(str$(bytesfloat))+result
	else
		result=format$(bytesfloat, "#.0")+result
	end if
	
	function=result
	
end function


sub KillParent()
	' Finds the PID of the parent cmd.exe instance and kills it, so that a new cmd can be created.
	' Not used anymore since v1.07 (we use an auxilliary batch file instead, to avoid this hack)
	
	local my_pid as long
	local parent_pid as long
	local dummy as long
	local i as long
	redim proc_list (0 to 0) as local ProcessEntry32
	
	' We start by getting this (jcd) process' PID.
	my_pid=GetCurrentProcessID
	dummy=procGetList(proc_list())
	
	' Cycle through all processes until we find this one. Then, we determine its parent.
	for i=0 to ubound(proc_list)
		if proc_list(i).th32ProcessID=my_pid then
			parent_pid=proc_list(i).th32ParentProcessID
			exit for
		end if
	next i
	
	'stdout "me="+format$(my_pid)+"; Parent="+format$(parent_pid)
	
	' Let's do the deed >:D
	procKill(parent_pid)
	
end sub
