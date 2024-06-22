SELECT 
    ordered_items, 
    total_price
FROM 
    pending_orders 
WHERE 
    order_id = $1;