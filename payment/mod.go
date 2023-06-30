package payment

// TODO: Use the valid request below to generate a payment link (ask copilot chat to write it for me)
// curl https://connect.squareupsandbox.com/v2/online-checkout/payment-links \
//   -X POST \
//   -H 'Square-Version: 2023-06-08' \
//   -H 'Authorization: Bearer foobar' \
//   -H 'Content-Type: application/json' \
//   -d '{
//     "idempotency_key": "6918319e-b9e3-4e2d-bea2-ea12e40926f0",
//     "checkout_options": {
//       "allow_tipping": false,
//       "ask_for_shipping_address": true,
//       "shipping_fee": {
//         "charge": {
//           "amount": 5,
//           "currency": "USD"
//         },
//         "name": "Flat Rate"
//       },
//       "redirect_url": "https://printing.focusandfilters.com/order_completed"
//     },
//     "order": {
//       "location_id": "LMMFFMJR68REF",
//       "customer_id": "foobar",
//       "line_items": [
//         {
//           "quantity": "1",
//           "base_price_money": {
//             "amount": 45,
//             "currency": "USD"
//           },
//           "item_type": "ITEM",
//           "name": "My Cool Print"
//         }
//       ],
//       "reference_id": "internal-id"
//     }
//   }'

// TODO: Fetch the order once the user gets to order completed and validate that it was paid
