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

class ConsoleClass
	instance cons_width as long
	instance cons_height as long
	
	class method create
		' Get current console dimensions at object creation
		
		console get screen to cons_height, cons_width
		cursor off
		page 1, 1	' We're not really using paging, save for the ShowHelp method.
		
	end method
	
	class method destroy
		' Turn the cursor on again
		
		cursor on
		
	end method
	
	class method ScrollHelp(byval scroll_pos as long, byval help_contents as string)
		' Private method, used for scrolling the help screen
		
		local i as long
		local num_lines as long
		local percent as double
		
		num_lines=parsecount(help_contents, $crlf)
		
		' Print help contents
		color 15, 1
		for i=5 to cons_height-3
			locate i, 5
			print lset$(parse$(help_contents, $crlf, scroll_pos+i-5), cons_width-8) 
		next i
		
		' Print line counter
		percent=min(1, (scroll_pos+cons_height-8)/num_lines)	' Compare 1 instead of 100 because the FORMAT$ mask used below multiplies by 100
		color 14, 1
		locate cons_height-2, 5
		print format$(percent, "0%")+"  "
		
	end method
	
	interface iConsole
		inherit iunknown
		
		' Getters and setters
		property get Width as long
			property=cons_width
		end property
		
		property get Height as long
			property=cons_height
		end property
		
		' Methods
		method DrawHeader(byval start_dir as string, byval oDir as iDirectoryInfo)
			' Draws all elements of the header except the current position within the listing (the offset)
			
			local buffer as string
			
			color 14, 1
			locate 1, 1
			if len(start_dir)>cons_width then
				' We have to shorten the shown directory.
				'print left$(start_dir, 10)+"[...]"+right$(start_dir, cons_width-15)	' Replaced by line below
				print left$(start_dir, cons_width-15)+"[...]"+right$(start_dir, 10)
			else
				' We pad the string to have the full width.
				print start_dir+space$(cons_width-len(start_dir))
			end if
			
			' Then some information about the current directory, and our position within it.
			color 15, 1
			locate 2, 1
			buffer=format$(oDir.NumDirs)+" directories, "+format$(oDir.NumEntries-oDir.NumDirs)+" files; "+_
				HumanFileSize(oDir.TotalSize)+" used."
			print buffer+space$(cons_width-len(buffer))
			
		end method
		
		method PrintCredits()
			stdout "JCD - Visual directory navigator ver. "+$version
			stdout "2010 by Jaime de los Hoyos M."
		end method
		
		method DrawOffset(byval current_position as long, byval num_entries as long, sort_mode as long)
			' Draws the currently selected item position/total entries in current dir
			
			local buffer as string
			
			if num_entries=0 then current_position=0
			
			select case sort_mode
				case 1
					buffer="[F2]=Toggle sort (Name Asc) "
				
				case 2
					buffer="[F2]=Toggle sort (Date Asc) "
				
				case 3
					buffer="[F2]=Toggle sort (Size Asc) "
				
				case 4
					buffer="[F2]=Toggle sort (Name Desc) "
				
				case 5
					buffer="[F2]=Toggle sort (Date Desc) "
				
				case 6
					buffer="[F2]=Toggle sort (Size Desc) "
					
			end select
			
			buffer="  "+buffer+format$(current_position)+"/"+format$(num_entries)
			
			color 14, 1
			locate 2, cons_width-len(buffer)-1
			print buffer
			
		end method
		
		method DrawFooter(byval search_str as string)
			' Draws the footer containing the program keys, and the current search pattern, if any
			
			local buffer as string
			
			buffer="[F1]=Show help [ESC]=Quit"
			color 14, 1
			locate cons_height, 1
			print buffer+space$(cons_width-len(buffer)-1);	' So that the screen does not scroll
			color 15, 1
			locate cons_height, 1
			print "[F1]";
			locate cons_height, 16
			print "[ESC]";
			
			color 15, 1
			locate cons_height, 30
			if search_str = "" then
				print "No search pattern";
			else
				print "Search pattern:";
				color 14, 1
				locate cons_height, 46
				print search_str;
			end if
			
		end method
		
		method DrawEntry(byval idx as long, byval row as long, byval highlight as long, byval oDir as iDirectoryInfo)
			' Draws an entry at the specified location, highlighted or normal.
			
			if oDir.NumEntries=0 then exit method	' Exit immediately if no entries in directory
			
			local buffer as string
			local buffer2 as string
			local i as long
			
			' Let's first check if it's a drive
			regexpr "^[A-Z]:$" in oDir.GetEntryName(idx) to i
			
			color 8, 0
			locate row, 1
			print space$(cons_width);
			locate row, 1
			if i=1 then
				' It's a drive
				buffer=oDir.GetEntryName(idx)
			else
				buffer=oDir.GetEntryAttr(idx)+" "+oDir.GetLastWriteTime(idx)+" "+oDir.GetEntryName(idx)
			end if
			
			if i=1 then
				' It's a drive
				buffer2=oDir.GetDriveAttr(idx)
				if highlight=0 then
					color 15, 0
				else
					color 0, 7
				end if
				print buffer;
				if highlight=0 then
					color 8, 0
				else
					color 0, 7
				end if
				print buffer2;
			elseif left$(buffer, 1)="d" then
				' It's a directory
				' Check whether the whole line will fit in the screen; if it will not, truncate part of the file name.
				' In that case, show the first ten characters of the entry, a [...], and then the remainder of the entry as the
				' screen width allows.
				if len(buffer)>cons_width then
					'buffer=left$(buffer, 39)+"[...]"+right$(buffer, (cons_width-len(buffer)-44))	' Replaced by line below
					buffer=left$(buffer, (cons_width-len(buffer)-15))+"[...]"+right$(buffer, 10)
				end if
				if highlight=0 then
					color 14, 0
				else
					color 0, 7
				end if
				print buffer;
			else
				' It's a file
				buffer2=" ("+HumanFileSize(oDir.GetEntrySize(idx))+")"
				' Check whether the whole line will fit in the screen; if it will not, truncate part of the file name.
				' In that case, show the first ten characters of the entry, a [...], and then the remainder of the entry as the
				' screen width allows.
				if len(buffer)+len(buffer2)>cons_width then
					'buffer=left$(buffer, 39)+"[...]"+right$(buffer, (cons_width-len(buffer)-(44+len(buffer2))))	' Replaced by line below
					buffer=left$(buffer, cons_width-len(buffer)-(15+len(buffer2)))+"[...]"+right$(buffer, 10)
				end if
				if highlight=0 then
					color 15, 0
				else
					color 0, 7
				end if
				print buffer;
				if highlight=0 then
					color 8, 0
				else
					color 0, 7
				end if
				print buffer2;
			end if
			
		end method
		
		method DrawScreen(byval offset as long, byval oDir as iDirectoryInfo)
			' Redraws all entries, from the specified offset, to fill the screen.
			' Does not highlight any entries.
			
			local i as long
			local entries as long
			
			' Check how many entries we have to fill.
			entries=oDir.NumEntries-offset
			entries=min(entries, cons_height-3)
			
			' Draw all entries
			if entries>0 then
				for i=1 to entries
					me.DrawEntry(i+offset, i+2, 0, oDir)
				next i
			end if
			
			' Blank all lines not used by entries
			if entries<cons_height-3 then
				color 8, 0
				for i=1 to cons_height-3-entries
					locate i+2+entries, 1
					print space$(cons_width);
				next i
			end if
			
		end method
		
		method ShowHelp()
			' Displays a help screen. When done, returns the console screen to the exact state it was in before invoking this screen.
			
			local i as long
			local j as long
			local help_contents as string
			local buffer as string
			local scroll_pos as long
			local keypressed as string
			local num_lines as long
			
			' Load help contents into our string, cutting long lines as necessary
			for i=1 to datacount
				buffer=read$(i)
				replace "<$ver>" with $version in buffer
				if len(buffer)>cons_width-8 then
					' We need to cut this line to fit inside the canvas
					do
						if len(buffer)<=cons_width-8 then
							help_contents+=buffer+$crlf
							exit do
						else
							j=instr(-1, left$(buffer, cons_width-8), " ")
							if j=0 or j>cons_width-8 then j=min(len(buffer), cons_width-8)	' This covers if no spaces were found
							help_contents+=left$(buffer, j)+$crlf								' to adequately cut the string.
							buffer=right$(buffer, len(buffer)-j)
						end if
					loop
				else
					help_contents+=buffer+$crlf
				end if
			next i
			num_lines=parsecount(help_contents, $crlf)
			
			pcopy 1, 8	' Save the console's current state to console page 8
			
			' Prepare the canvas
			for i=4 to cons_height-2
				locate i, 4
				color 14, 1
				print space$(cons_width-6);
				color 14, 8
				if i=4 then		' Right border "shadow"
					print ""
				else
					color 0, 8, 1
					print ""
				end if
			next i
			locate cons_height-1, 6		' Lower border "shadow"
			color 0, 8, (cons_width-7)
			
			' Print help's help
			color 14, 1
			locate cons_height, 1
			print lset$("[Up/Down]=Scroll [ESC] or [F1]=Exit help", cons_width-1);
			color 15, 1
			locate cons_height, 1
			print "[Up/Down]";
			locate cons_height, 18
			print "[ESC]";
			locate cons_height, 27
			print "[F1]";
			
			' Scroll cycle
			scroll_pos=1
			
			me.ScrollHelp(scroll_pos, help_contents)
			do
				keypressed=waitkey$
				select case keypressed
					case chr$(27), chr$(0)+chr$(59)	' Esc or F1 was pressed
						exit do
					
					case chr$(0)+chr$(72)	' Arrow up
						if scroll_pos>1 then
							decr scroll_pos
							me.ScrollHelp(scroll_pos, help_contents)
						end if
					
					case chr$(0)+chr$(80)	' Arrow down
						if scroll_pos<=num_lines-(cons_height-7) then
							incr scroll_pos
							me.ScrollHelp(scroll_pos, help_contents)
						end if
					
				end select
			loop
			
			pcopy 8, 1	' Restore the console exactly how it was before we called this method.
			
			' Strings for the help screen
			data "JVA Directory Navigator version <$ver>"
			data "by Jaime de los Hoyos M. (jdeloshoyos@gmail.com)"
			data "------------------------------------------------"
			data ""
			data "Commands:"
			data "[Up/Down/PgUp/PgDn]=Navigate"
			data "[Enter]=Select file or open directory"
			data "[Backspace]=Go up one level"
			data "[ESC]=Quit"
			data "[F1]=Toggle this help screen"
			data "[F2]=Toggle sorting between sort by name, last write time, or file size"
			data "[Tab]=Cycle through search results"
			data ""
			data "Search entries:"
			data "---------------"
			data ""
			data "A search is started by typing a character sequence to look for anywhere within a file or directory entry. If a match is found, the program will 'jump to' and highlight the first entry that matches the typed characters."
			data ""
			data "If more than one entry matches the current search, pressing [Tab] will cycle between all results."
			data ""
			data "To start a new search, simply wait a few seconds and type a new search pattern."
			data ""
			data "Entry attributes:"
			data "-----------------"
			data ""
			data "Each listed entry in a directory (both files and directories) has a list of attributes at the beginning of the line, in the form of a list of characters in a line of dashes. If a letter is shown instead of a dash, that entry has that attribute."
			data ""
			data "The meaning of each attribute is as follows:"
			data ""
			data "d = Directory"
			data "l = Reparse point/symbolic link"
			data "r = Read only"
			data "h = Hidden"
			data "s = System file"
			data "a = Archive"
			data "c = Compressed"
			data "i = Not content indexed"
			data "e = Encrypted"
			
		end method
		
	end interface
	
end class
