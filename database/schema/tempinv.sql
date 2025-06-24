-- ----------------------------------------------------------
-- MDB Tools - A library for reading MS Access database files
-- Copyright (C) 2000-2011 Brian Bruns and others.
-- Files in libmdb are licensed under LGPL and the utilities under
-- the GPL, see COPYING.LIB and COPYING files respectively.
-- Check out http://mdbtools.sourceforge.net
-- ----------------------------------------------------------

-- That file uses encoding UTF-8

CREATE TABLE [tempinv]
 (
	[id]			Long Integer, 
	[username]			Text (50), 
	[wkorder]			Text (50), 
	[custid]			Long Integer, 
	[customer]			Text (50), 
	[joints]			Long Integer, 
	[rack]			Text (50), 
	[size]			Text (50), 
	[weight]			Text (50), 
	[grade]			Text (50), 
	[connection]			Text (50), 
	[ctd]			Boolean NOT NULL, 
	[wstring]			Boolean NOT NULL, 
	[swgcc]			Text (50), 
	[color]			Text (50), 
	[customerpo]			Text (50), 
	[fletcher]			Text (50), 
	[datein]			DateTime, 
	[dateout]			DateTime, 
	[wellin]			Text (50), 
	[leasein]			Text (50), 
	[wellout]			Text (50), 
	[leaseout]			Text (50), 
	[trucking]			Text (50), 
	[trailer]			Text (50), 
	[LOCATION]			Text (50), 
	[notes]			Memo/Hyperlink (255), 
	[pcode]			Text (50), 
	[cn]			Long Integer
);


