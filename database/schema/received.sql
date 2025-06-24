-- ----------------------------------------------------------
-- MDB Tools - A library for reading MS Access database files
-- Copyright (C) 2000-2011 Brian Bruns and others.
-- Files in libmdb are licensed under LGPL and the utilities under
-- the GPL, see COPYING.LIB and COPYING files respectively.
-- Check out http://mdbtools.sourceforge.net
-- ----------------------------------------------------------

-- That file uses encoding UTF-8

CREATE TABLE [RECEIVED]
 (
	[ID]			Long Integer, 
	[WKORDER]			Text (50), 
	[CUSTID]			Long Integer, 
	[CUSTOMER]			Text (50), 
	[JOINTS]			Long Integer, 
	[RACK]			Text (50), 
	[SIZEID]			Long Integer, 
	[SIZE]			Text (50), 
	[WEIGHT]			Text (50), 
	[GRADE]			Text (50), 
	[CONNECTION]			Text (50), 
	[CTD]			Boolean NOT NULL, 
	[WSTRING]			Boolean NOT NULL, 
	[WELL]			Text (50), 
	[LEASE]			Text (50), 
	[ORDEREDBY]			Text (50), 
	[NOTES]			Memo/Hyperlink (255), 
	[CUSTOMERPO]			Text (50), 
	[DATERECVD]			DateTime, 
	[BACKGROUND]			Text (50), 
	[NORM]			Text (50), 
	[SERVICES]			Text (50), 
	[BILLTOID]			Text (50), 
	[ENTEREDBY]			Text (50), 
	[WHEN1]			DateTime, 
	[TRUCKING]			Text (50), 
	[TRAILER]			Text (50), 
	[inproduction]			DateTime, 
	[inspected]			DateTime, 
	[threading]			DateTime, 
	[straighten]			Boolean NOT NULL, 
	[excess]			Boolean NOT NULL, 
	[COMPLETE]			Boolean NOT NULL, 
	[inspectedby]			Text (50), 
	[updatedby]			Text (50), 
	[when2]			DateTime, 
	[deleted]			Boolean NOT NULL
);


