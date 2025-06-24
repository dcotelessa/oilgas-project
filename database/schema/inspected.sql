-- ----------------------------------------------------------
-- MDB Tools - A library for reading MS Access database files
-- Copyright (C) 2000-2011 Brian Bruns and others.
-- Files in libmdb are licensed under LGPL and the utilities under
-- the GPL, see COPYING.LIB and COPYING files respectively.
-- Check out http://mdbtools.sourceforge.net
-- ----------------------------------------------------------

-- That file uses encoding UTF-8

CREATE TABLE [inspected]
 (
	[id]			Long Integer, 
	[username]			Text (50), 
	[wkorder]			Text (50), 
	[color]			Text (50), 
	[joints]			Long Integer, 
	[accept]			Long Integer, 
	[reject]			Long Integer, 
	[pin]			Long Integer, 
	[cplg]			Long Integer, 
	[pc]			Long Integer, 
	[complete]			Boolean NOT NULL, 
	[rack]			Text (50), 
	[rep_pin]			Long Integer, 
	[rep_cplg]			Long Integer, 
	[rep_pc]			Long Integer, 
	[deleted]			Boolean NOT NULL, 
	[cn]			Long Integer
);


