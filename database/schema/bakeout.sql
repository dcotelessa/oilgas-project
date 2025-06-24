-- ----------------------------------------------------------
-- MDB Tools - A library for reading MS Access database files
-- Copyright (C) 2000-2011 Brian Bruns and others.
-- Files in libmdb are licensed under LGPL and the utilities under
-- the GPL, see COPYING.LIB and COPYING files respectively.
-- Check out http://mdbtools.sourceforge.net
-- ----------------------------------------------------------

-- That file uses encoding UTF-8

CREATE TABLE [bakeout]
 (
	[id]			Long Integer, 
	[fletcher]			Text (50), 
	[joints]			Long Integer, 
	[color]			Text (50), 
	[size]			Text (50), 
	[weight]			Text (50), 
	[grade]			Text (50), 
	[connection]			Text (50), 
	[ctd]			Boolean NOT NULL, 
	[swgcc]			Text (50), 
	[custid]			Long Integer, 
	[accept]			Long Integer, 
	[reject]			Long Integer, 
	[pin]			Long Integer, 
	[cplg]			Long Integer, 
	[pc]			Long Integer, 
	[trucking]			Text (50), 
	[trailer]			Text (50), 
	[datein]			DateTime, 
	[cn]			Long Integer
);


