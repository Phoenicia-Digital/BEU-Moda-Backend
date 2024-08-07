SELECT 
    user_id,
    non_user_full_name,
    non_user_billing_address,
    non_user_phone_number,
    ordered_items, 
    total_price, 
    order_time 
FROM 
    pending_orders
WHERE
    order_id = $1;