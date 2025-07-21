<cfquery name="getCustomers" datasource="inventory">
    SELECT CustID, CustName, BillAddr 
    FROM Customers 
    WHERE IsDeleted = 0
    ORDER BY CustName
</cfquery>

<cfquery name="getInventory" datasource="inventory">
    SELECT i.WkOrder, i.Joints, i.Size, i.Grade, c.CustName
    FROM Inventory i
    INNER JOIN Customers c ON i.CustID = c.CustID
    WHERE i.DateIn >= #CreateDate(2024,1,1)#
</cfquery>
