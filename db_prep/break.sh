#!/usr/bin/env bash

# Check if mdbtools are installed
command -v mdb-tables >/dev/null 2>&1 || { echo >&2 "I require mdb-tables but it's not installed. Aborting."; exit 1; }
command -v mdb-export >/dev/null 2>&1 || { echo >&2 "I require mdb-export but it's not installed. Aborting."; exit 1; }

# Define the MDB database file
fullfilename="petros-bk.mdb"

# Create a directory to store the CSV files
filename=$(basename "$fullfilename")
dbname=${filename%.*}
mkdir -p "$dbname"

# Loop through each table and export it
IFS=$'\n' # Set Internal Field Separator to newline to handle table names with spaces
for table in $(mdb-tables -1 "$fullfilename"); do
    echo "Exporting table: $table"
    mdb-export "$fullfilename" "$table" > "$dbname/$table.csv"
done

echo "All tables exported to the '$dbname' directory."
