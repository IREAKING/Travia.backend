# 🚀 Stripe Setup Guide cho Travia

## Bước 1: Tạo Stripe Account

1. Truy cập https://dashboard.stripe.com/register
2. Điền thông tin:
   - Email
   - Full name
   - Country: Vietnam
   - Company name: Travia Travel

3. Xác thực email

---

## Bước 2: Lấy API Keys

### Test Mode (Development)

1. Vào https://dashboard.stripe.com/test/apikeys
2. Copy 2 keys sau:

```bash
# Publishable key (pk_test_...)
pk_test_51xxxxx

# Secret key (sk_test_...) - Click "Reveal live key"
sk_test_51xxxxx
```

3. Thêm vào `env/.env`:
```bash
STRIPE_SECRET_KEY=sk_test_51xxxxx
STRIPE_PUBLISHABLE_KEY=pk_test_51xxxxx
```

### Live Mode (Production)

⚠️ **Chỉ dùng khi đã test kỹ và sẵn sàng nhận tiền thật!**

1. Complete business verification
2. Add bank account details
3. Switch to "Live mode" trong dashboard
4. Get live keys tương tự test keys

---

## Bước 3: Setup Webhook

### Local Development (với Stripe CLI)

```bash
# 1. Install Stripe CLI
brew install stripe/stripe-cli/stripe

# 2. Login
stripe login

# 3. Forward webhooks to local server
stripe listen --forward-to http://localhost:8080/api/payment/webhook

# Output sẽ có webhook signing secret:
# whsec_xxxxxxxxxxxxx
```

Copy webhook secret vào `.env`:
```bash
STRIPE_WEBHOOK_SECRET=whsec_xxxxxxxxxxxxx
```

### Production (Stripe Dashboard)

1. Vào https://dashboard.stripe.com/test/webhooks
2. Click "Add endpoint"
3. Nhập URL:
   ```
   https://your-domain.com/api/payment/webhook
   ```
4. Chọn events:
   - `payment_intent.succeeded`
   - `payment_intent.payment_failed`
   - `charge.refunded`
5. Click "Add endpoint"
6. Copy webhook signing secret
7. Add vào production `.env`

---

## Bước 4: Chạy Migration

```bash
# Connect to database
psql -U postgres -d travia_db

# Run migration
\i db/migration/002_add_payments.sql

# Verify tables created
\dt thanh_toan
\dt hoan_tien
\dt stripe_webhook_log
```

---

## Bước 5: Install Dependencies

```bash
cd /path/to/Travia.backend

# Install Stripe Go SDK
go get github.com/stripe/stripe-go/v79

# Tidy dependencies
go mod tidy
```

---

## Bước 6: Test Implementation

### 1. Start Server

```bash
# Terminal 1: Start Stripe webhook forwarding
stripe listen --forward-to http://localhost:8080/api/payment/webhook

# Terminal 2: Start backend
go run main.go
```

### 2. Test API

```bash
# Get Stripe config
curl http://localhost:8080/api/payment/config

# Should return:
{
  "publishable_key": "pk_test_...",
  "currency": "usd"
}
```

### 3. Test Payment Flow (với Postman hoặc HTTP file)

```http
POST http://localhost:8080/api/payment/create-intent
Content-Type: application/json
Authorization: Bearer your_token

{
  "booking_id": 1,
  "amount": 100.00,
  "currency": "usd"
}
```

### 4. Test với Stripe CLI

```bash
# Trigger test webhook
stripe trigger payment_intent.succeeded

# Check logs
stripe listen --print-json
```

---

## Bước 7: Frontend Integration

### Install Stripe.js

```bash
npm install @stripe/stripe-js @stripe/react-stripe-js
```

### Basic Integration

```javascript
import { loadStripe } from '@stripe/stripe-js';
import { Elements, CardElement, useStripe, useElements } from '@stripe/react-stripe-js';

// Initialize Stripe
const stripePromise = loadStripe('pk_test_...');

function CheckoutForm() {
  const stripe = useStripe();
  const elements = useElements();

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    // 1. Create payment intent from backend
    const response = await fetch('/api/payment/create-intent', {
      method: 'POST',
      headers: { 
        'Content-Type': 'application/json',
        'Authorization': 'Bearer ' + token
      },
      body: JSON.stringify({
        booking_id: 1,
        amount: 100,
        currency: 'usd'
      })
    });
    
    const { client_secret } = await response.json();
    
    // 2. Confirm payment
    const result = await stripe.confirmCardPayment(client_secret, {
      payment_method: {
        card: elements.getElement(CardElement)
      }
    });
    
    if (result.error) {
      alert('Payment failed: ' + result.error.message);
    } else {
      alert('Payment successful!');
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <CardElement />
      <button type="submit">Pay</button>
    </form>
  );
}

export default function App() {
  return (
    <Elements stripe={stripePromise}>
      <CheckoutForm />
    </Elements>
  );
}
```

---

## Bước 8: Testing với Test Cards

### Success Card
```
Card: 4242 4242 4242 4242
Expiry: 12/34 (any future date)
CVC: 123 (any 3 digits)
ZIP: 12345
```

### 3D Secure (requires authentication)
```
Card: 4000 0025 0000 3155
```

### Declined Cards
```
# Insufficient funds
Card: 4000 0000 0000 9995

# Expired card
Card: 4000 0000 0000 0069

# Incorrect CVC
Card: 4000 0000 0000 0127
```

---

## Bước 9: Monitoring & Logs

### Stripe Dashboard

1. **Payments:** https://dashboard.stripe.com/test/payments
2. **Webhooks:** https://dashboard.stripe.com/test/webhooks
3. **Logs:** https://dashboard.stripe.com/test/logs

### Backend Logs

```bash
# View payment logs
tail -f logs/payment.log

# Query database
psql -U postgres -d travia_db -c "SELECT * FROM thanh_toan ORDER BY ngay_tao DESC LIMIT 10;"

# Check webhook logs
psql -U postgres -d travia_db -c "SELECT * FROM stripe_webhook_log ORDER BY ngay_nhan DESC LIMIT 10;"
```

---

## Bước 10: Go Live Checklist

### Before Production:

- [ ] Test all payment flows thoroughly
- [ ] Test refund process
- [ ] Test webhook handling
- [ ] Review error handling
- [ ] Setup proper logging
- [ ] Configure monitoring/alerts
- [ ] Complete Stripe account verification
- [ ] Add production bank account
- [ ] Switch to live API keys
- [ ] Update webhook URLs to production
- [ ] Test with small real transactions
- [ ] Setup email notifications
- [ ] Document payment procedures
- [ ] Train support team

### Security Checklist:

- [ ] Never commit API keys to Git
- [ ] Use environment variables
- [ ] Enable HTTPS only
- [ ] Validate webhook signatures
- [ ] Implement rate limiting
- [ ] Log all transactions
- [ ] Regular security audits

---

## 💰 Fees & Pricing

### Stripe Fees (as of 2025):

| Transaction Type | Fee |
|-----------------|-----|
| Vietnam cards | 3.9% + 30¢ |
| International cards | 3.9% + 30¢ |
| Currency conversion | +1% |
| Dispute/chargeback | $15 |

### Monthly Costs Example:

```
100 transactions/month
Average: $200/transaction
Total volume: $20,000

Fees: $20,000 × 3.9% + (100 × $0.30) = $810
Net: $19,190

Percentage: 4.05%
```

---

## 🆘 Troubleshooting

### Issue: "API key invalid"

**Solution:**
- Check `.env` file has correct keys
- Verify no extra spaces in keys
- Ensure using test keys for test mode

### Issue: "Webhook signature failed"

**Solution:**
- Use Stripe CLI for local testing
- Verify `STRIPE_WEBHOOK_SECRET` is correct
- Check webhook endpoint URL is correct

### Issue: "Payment failed with 3D Secure"

**Solution:**
- This is normal for some cards
- Frontend must handle authentication
- See Stripe.js documentation for 3DS

---

## 📚 Additional Resources

- **Stripe Dashboard:** https://dashboard.stripe.com/
- **API Docs:** https://stripe.com/docs/api
- **Testing Guide:** https://stripe.com/docs/testing
- **Integration Builder:** https://stripe.com/docs/payments/quickstart
- **Support:** https://support.stripe.com/

---

## 🎉 Done!

Stripe payment đã sẵn sàng cho Travia! 

**Next steps:**
1. Test payment flow thoroughly
2. Integrate frontend
3. Add analytics tracking
4. Setup monitoring alerts

**Questions?** Check documentation hoặc Stripe support.

---

**Created:** October 2025  
**Last Updated:** October 2025

