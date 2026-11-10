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

class DirectoryInfo
	instance num_entries as long
	instance dir_entries() as DirData	' This UDT is built into PB
	instance max_entries as long
	instance num_dirs as long
	instance total_bytes as quad
	instance sort_mode as long
	
	class method create
		' Set the array to a fixed initial size.
		' It will grow as needed.
		
		me.IncreaseArraySize
		
		' Set initial sorting mode to sort by name
		sort_mode=1
		
	end method
	
	class method IncreaseArraySize
		' Makes the entries array grow.
		
		max_entries=max_entries+5000
		redim preserve dir_entries(max_entries) as instance DirData
		
	end method
	
	interface iDirectoryInfo
		inherit iunknown
		
		' Getters and setters
		property get NumEntries as long
			property=num_entries
		end property
		
		property get NumDirs as long
			property=num_dirs
		end property
		
		property get TotalSize as quad
			property=total_bytes
		end property
		
		property get SortMode as long
			property=sort_mode
		end property
		
		property set SortMode (byval i as long)
			' Sets the sorting mode.
			' 1 = Sort by file name (the default) ascending
			' 2 = Sort by last write time ascending
			' 3 = Sort by file size ascending
			' 4 = Sort by file name descending
			' 5 = Sort by last write time descending
			' 6 = Sort by file size descending
			
			if i>0 and i<7 then sort_mode=i
			
		end property
		
		' Methods
		method ScanDir(byval start_dir as string)
			' Fills the dir_entries array with the... dir entries.
			' It also sets the num_entries and num_dirs variables.
			
			total_bytes=0
			
			local i as long
			local buffer as string

			' v1.11: We retrieve everything, files AND subdirs (including hidden items).
			num_entries = 0
			num_dirs = 0
			buffer=dir$(start_dir+"\*", %NORMAL + %HIDDEN + %SYSTEM + %SUBDIR, to dir_entries(i))
			
			if buffer<>"" then
				' Check if it's a directory
				if bit(dir_entries(i).FileAttributes, 4) then incr num_dirs
				
				incr i
				total_bytes=total_bytes+me.GetEntrySize(i)
				if i>max_entries then me.IncreaseArraySize
				do
					buffer=dir$(next to dir_entries(i))
					if buffer="" then
						exit do
					else
						' Check if it's a directory
						if bit(dir_entries(i).FileAttributes, 4) then incr num_dirs
						
						incr i
						total_bytes=total_bytes+me.GetEntrySize(i)
						if i>max_entries then me.IncreaseArraySize
					end if
				loop
			end if
			
			' Done. i contains the actual number of entries for start_dir (regardless of how many entries are loaded into the array).
			num_entries=i
			dir$ close
			
			' Sort entries according to the previously set sorting method
			' 1 = Sort by file name (the default)
			' 2 = Sort by last write time
			' 3 = Sort by file size
			select case sort_mode
				case 1
					' Sort by name, descending. Dirs always go first. This is the default sorting mode.
					array sort dir_entries(0) for num_entries, using DirectoryInfo_class_SortByNameAsc()
				
				case 2
					' Sort by last write time, ascending. Dirs always go first.
					array sort dir_entries(0) for num_entries, using DirectoryInfo_class_SortByDateAsc()
				
				case 3
					' Sort by file size, ascending. Dirs always go first.
					array sort dir_entries(0) for num_entries, using DirectoryInfo_class_SortBySizeAsc()
				
				case 4
					' Sort by name, descending. Dirs always go first.
					array sort dir_entries(0) for num_entries, using DirectoryInfo_class_SortByNameDesc()
				
				case 5
					' Sort by last write time, descending. Dirs always go first.
					array sort dir_entries(0) for num_entries, using DirectoryInfo_class_SortByDateDesc()
				
				case 6
					' Sort by file size, descending. Dirs always go first.
					array sort dir_entries(0) for num_entries, using DirectoryInfo_class_SortBySizeDesc()
					
			end select
			
		end method
		
		method GetEntryName(byval idx as long) as string
			' Returns the name of the specified item
			
			method=dir_entries(idx-1).FileName	' The array is zero-based
			
		end method
		
		method GetLastWriteTime(byval idx as long) as string
			' Returns the last write time of the specified item
			
			local lft as FileTime
			local st AS SystemTime
			local buffer as string
			local i as long
			
			' This converts the QUAD returned in the DirData UDT to a
			' human readable date and time. Credit goes to José Roca
			FileTimeToLocalFileTime(BYVAL VarPTR(dir_entries(idx-1).LastWriteTime), lft)
			FileTimeToSystemTime(lft, st)
			buffer = Format$(st.wYear) + "/" + Format$(st.wMonth, "00") + "/" + Format$(st.wDay, "00") + " " + _
				Format$(st.wHour, "00") + ":" + Format$(st.wMinute, "00") + ":"+Format$(st.wSecond, "00")
				
			method=buffer
			
		end method
		
		method GetEntrySize(byval idx as long) as quad
			' Returns the size, in bytes, of the specified item.
			
			method=mak(quad, dir_entries(idx-1).FileSizeLow, dir_entries(idx-1).FileSizeHigh)	' The array is zero-based
			
		end method
		
		method GetEntryAttr(byval idx as long) as string
			' Returns a string mask, containing the attributes of the specified item.
			' Following is the meaning of most of the important bits contained in the FileAttributes (from MSDN)
			' http://msdn.microsoft.com/en-us/library/gg258117%28v=VS.85%29.aspx
			' bit	dec	hex	label											letter
			' 0	1		1		FILE_ATTRIBUTE_READONLY					r
			' 1	2		2		FILE_ATTRIBUTE_HIDDEN					h
			' 2	4		4		FILE_ATTRIBUTE_SYSTEM					s
			' 3	8		8		
			' 4	16		10		FILE_ATTRIBUTE_DIRECTORY				d
			' 5	32		20		FILE_ATTRIBUTE_ARCHIVE					a
			' 6	64		40		
			' 7	128	80		FILE_ATTRIBUTE_NORMAL	
			' 8	256	100		
			' 9	512	200		
			' 10	1024	400	FILE_ATTRIBUTE_REPARSE_POINT        l	
			' 11	2048	800	FILE_ATTRIBUTE_COMPRESSED				c
			' 12	4096	1000		
			' 13	8192	2000	FILE_ATTRIBUTE_NOT_CONTENT_INDEXED	i
			' 14	16384	4000	FILE_ATTRIBUTE_ENCRYPTED				e
			' 15	32768	8000
			'
			' I will use this order for the flags (somewhat arbitrary, but I'll put the dir flag first :)
			' drhsacie
			' Flags not set will be replaced by dashes (a la unix' ls)
			
			local attr as dword
			local attr_mask as string
			
			attr_mask="---------"
			
			attr=dir_entries(idx-1).FileAttributes
			
			if bit(attr, 4) then asc(attr_mask, 1)=asc("d")
			if bit(attr, 10) then asc(attr_mask, 2)=asc("l")
			if bit(attr, 0) then asc(attr_mask, 3)=asc("r")
			if bit(attr, 1) then asc(attr_mask, 4)=asc("h")
			if bit(attr, 2) then asc(attr_mask, 5)=asc("s")
			if bit(attr, 5) then asc(attr_mask, 6)=asc("a")
			if bit(attr, 11) then asc(attr_mask, 7)=asc("c")
			if bit(attr, 13) then asc(attr_mask, 8)=asc("i")
			if bit(attr, 14) then asc(attr_mask, 9)=asc("e")
			
			method=attr_mask
			
		end method
		
		method GetDriveAttr(byval idx as long) as string
			' Returns what kind of drive it is
			
			local attr_mask as string
			
			select case dir_entries(idx-1).FileAttributes
				case 0
					attr_mask=" (Unknown)"

				case 2
					attr_mask=" (Removable drive)"
					
				case 3
					attr_mask=" (Hard drive)"
					
				case 4
					attr_mask=" (Network drive)"
					
				case 5
					attr_mask=" (CD/DVD)"
					
				case 6
					attr_mask=" (RAM drive)"
					
			end select
			
			method=attr_mask
			
		end method
		
		method ScanDrives()
			' Populates the array with the list of available drives.
			
			local dirmask as string
			local drivetype as long
			local i as long
			
			total_bytes=0
			num_entries=0
			num_dirs=0
			
			for i=asc("C") to asc("Z")
				dirmask=CHR$(i)+":"
				drivetype=GetDriveType(dirmask+$NUL)
				
				if drivetype<>1 then
					' A drivetype = 1 means it is NOT a valid drive connected to the system.
					dir_entries(num_entries).FileName=dirmask
					dir_entries(num_entries).FileAttributes=drivetype	' We store here the drive type.
					incr num_entries
					incr num_dirs
				end if
				
			next i
			
		end method
		
		method SearchEntry(byval search_str as string, byval idx as long) as long
			' Search within the current directory entries the (idx)st match that starts with search_str,
			' and returns its index number (1-based).
			' The search is case-insensitive.
			' If there are less matches than idx, then the last match is returned.
			' If idx = 0, then the value returned is a COUNT of all matches found, instead of the index of a match.
			
			local i as long
			local search_str_len as long
			local matches as long
			local last_match as long
			
			if num_entries=0 then exit method
			
			search_str_len=len(search_str)	' Precalculate out of the for/next
			
			if idx=0 then
				for i=0 to num_entries-1
					'if lcase$(left$(dir_entries(i).FileName, search_str_len))=lcase$(search_str) then incr matches
					if instr(lcase$(dir_entries(i).FileName), lcase$(search_str)) then incr matches
				next i
				
				method=matches
			else
				for i=0 to num_entries-1
					'if lcase$(left$(dir_entries(i).FileName, search_str_len))=lcase$(search_str) then
					if instr(lcase$(dir_entries(i).FileName), lcase$(search_str)) then
						' Match found!
						incr matches
						last_match=i
						if matches=idx then exit for
					end if
				next i
				
				if matches>0 then method=last_match+1
			end if
			
		end method
		
	end interface
	
end class

' Auxiliary functions for custom ARRAY SORTing of the directory listings
' ----------------------------------------------------------------------
function DirectoryInfo_class_SortByNameAsc(byref p1 as DirData, byref p2 as DirData) as long
	function=DirectoryInfo_class_SortByName(p1, p2, 1)
end function

function DirectoryInfo_class_SortByDateAsc(byref p1 as DirData, byref p2 as DirData) as long
	function=DirectoryInfo_class_SortByDate(p1, p2, 1)
end function

function DirectoryInfo_class_SortBySizeAsc(byref p1 as DirData, byref p2 as DirData) as long
	function=DirectoryInfo_class_SortBySize(p1, p2, 1)
end function

function DirectoryInfo_class_SortByNameDesc(byref p1 as DirData, byref p2 as DirData) as long
	function=DirectoryInfo_class_SortByName(p1, p2, -1)
end function

function DirectoryInfo_class_SortByDateDesc(byref p1 as DirData, byref p2 as DirData) as long
	function=DirectoryInfo_class_SortByDate(p1, p2, -1)
end function

function DirectoryInfo_class_SortBySizeDesc(byref p1 as DirData, byref p2 as DirData) as long
	function=DirectoryInfo_class_SortBySize(p1, p2, -1)
end function

function DirectoryInfo_class_SortByName(p1 as DirData, p2 as DirData, order as long) as long
	local attr1 as dword
	local attr2 as dword
	local isdir1 as long
	local isdir2 as long
	
	' First we'l get all directories first.
	
	attr1=p1.FileAttributes
	attr2=p2.FileAttributes
	
	if bit(attr1, 4) then isdir1=1
	if bit(attr2, 4) then isdir2=1
	
	if isdir1>isdir2 then
		function=-1
		exit function
	end if
	
	if isdir1<isdir2 then
		function=1
		exit function
	end if
	
	' We'll sort by file names
	if ucase$(p1.FileName)<ucase$(p2.FileName) then
		function=-1 * order
		exit function
	end if
	
	if ucase$(p1.FileName)>ucase$(p2.FileName) then
		function=1 * order
		exit function
	end if
	
end function

function DirectoryInfo_class_SortByDate(p1 as DirData, p2 as DirData, order as long) as long
	
	local attr1 as dword
	local attr2 as dword
	local isdir1 as long
	local isdir2 as long
	
	' First we'l get all directories first.
	
	attr1=p1.FileAttributes
	attr2=p2.FileAttributes
	
	if bit(attr1, 4) then isdir1=1
	if bit(attr2, 4) then isdir2=1
	
	if isdir1>isdir2 then
		function=-1
		exit function
	end if
	
	if isdir1<isdir2 then
		function=1
		exit function
	end if
	
	' Then, the date comparison
	if p1.LastWriteTime<p2.LastWriteTime then
		function=-1 * order
		exit function
	end if
	
	if p1.LastWriteTime>p2.LastWriteTime then
		function=1 * order
		exit function
	end if
	
	' Only if dates are equal, we'll sort by file names
	if ucase$(p1.FileName)<ucase$(p2.FileName) then
		function=-1
		exit function
	end if
	
	if ucase$(p1.FileName)>ucase$(p2.FileName) then
		function=1
		exit function
	end if
	
end function

function DirectoryInfo_class_SortBySize(p1 as DirData, p2 as DirData, order as long) as long
	
	local attr1 as dword
	local attr2 as dword
	local isdir1 as long
	local isdir2 as long
	local size1 as quad
	local size2 as quad
	
	' First we'll get all directories
	
	attr1=p1.FileAttributes
	attr2=p2.FileAttributes
	
	if bit(attr1, 4) then isdir1=1
	if bit(attr2, 4) then isdir2=1
	
	if isdir1>isdir2 then
		function=-1
		exit function
	end if
	
	if isdir1<isdir2 then
		function=1
		exit function
	end if
	
	' Then, the size comparison
	size1=mak(quad, p1.FileSizeLow, p1.FileSizeHigh)
	size2=mak(quad, p2.FileSizeLow, p2.FileSizeHigh)
	
	if size1<size2 then
		function=-1 * order
		exit function
	end if
	
	if size1>size2 then
		function=1 * order
		exit function
	end if
	
	' Only if sizes are equal, we'll sort by file names
	if ucase$(p1.FileName)<ucase$(p2.FileName) then
		function=-1
		exit function
	end if
	
	if ucase$(p1.FileName)>ucase$(p2.FileName) then
		function=1
		exit function
	end if
	
end function