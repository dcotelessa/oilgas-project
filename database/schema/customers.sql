-- ----------------------------------------------------------
-- MDB Tools - A library for reading MS Access database files
-- Copyright (C) 2000-2011 Brian Bruns and others.
-- Files in libmdb are licensed under LGPL and the utilities under
-- the GPL, see COPYING.LIB and COPYING files respectively.
-- Check out http://mdbtools.sourceforge.net
-- ----------------------------------------------------------

-- That file uses encoding UTF-8

CREATE TABLE [customers]
 (
	[custid]			Long Integer, 
	[customer]			Text (50), 
	[billingaddress]			Text (50), 
	[billingcity]			Text (50), 
	[billingstate]			Text (50), 
	[billingzipcode]			Text (50), 
	[contact]			Text (50), 
	[phone]			Text (50), 
	[fax]			Text (50), 
	[email]			Text (50), 
	[color1]			Text (50), 
	[color2]			Text (50), 
	[color3]			Text (50), 
	[color4]			Text (50), 
	[color5]			Text (50), 
	[loss1]			Text (50), 
	[loss2]			Text (50), 
	[loss3]			Text (50), 
	[loss4]			Text (50), 
	[loss5]			Text (50), 
	[wscolor1]			Text (50), 
	[wscolor2]			Text (50), 
	[wscolor3]			Text (50), 
	[wscolor4]			Text (50), 
	[wscolor5]			Text (50), 
	[wsloss1]			Text (50), 
	[wsloss2]			Text (50), 
	[wsloss3]			Text (50), 
	[wsloss4]			Text (50), 
	[wsloss5]			Text (50), 
	[deleted]			Boolean NOT NULL
);


