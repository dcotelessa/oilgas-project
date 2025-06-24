-- ----------------------------------------------------------
-- MDB Tools - A library for reading MS Access database files
-- Copyright (C) 2000-2011 Brian Bruns and others.
-- Files in libmdb are licensed under LGPL and the utilities under
-- the GPL, see COPYING.LIB and COPYING files respectively.
-- Check out http://mdbtools.sourceforge.net
-- ----------------------------------------------------------

-- That file uses encoding UTF-8

CREATE TABLE [users]
 (
	[userid]			Long Integer, 
	[username]			Text (12), 
	[password]			Text (12), 
	[access]			Long Integer, 
	[fullname]			Text (50), 
	[email]			Text (50)
);


