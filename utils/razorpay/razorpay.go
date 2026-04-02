package razorpay

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"voyagear/config"

	rzp "github.com/razorpay/razorpay-go"
)

type RazorpayClient struct {
	Client    *rzp.Client
	KeyID     string
	KeySecret string
}

func NewRazorpayClient(cfg *config.RazorpayConfig) *RazorpayClient {
	client := rzp.NewClient(cfg.KeyID, cfg.KeySecret)
	return &RazorpayClient{
		Client:    client,
		KeyID:     cfg.KeyID,
		KeySecret: cfg.KeySecret,
	}
}

// Create order creates new razorpay order
func (r *RazorpayClient) CreateOrder(amount float64, orderID string) (string, error) {

	amountInPaise := int(amount * 100) // razorpay amount is in paise (1 rupee = 100 paise)

	data := map[string]interface{}{
		"amount":   amountInPaise,
		"currency": "INR",
		"receipt":  fmt.Sprintf("order_%d", orderID),
		"notes": map[string]interface{}{
			"order_id": orderID,
		},
	}

	body, err := r.Client.Order.Create(data, nil)
	if err != nil {
		return "", fmt.Errorf("Failed to create razorpay order: %w", err)
	}

	razorpayOrderID, ok := body["id"].(string)
	if !ok {
		return "", fmt.Errorf("Invalid razorpay response")
	}

	return razorpayOrderID, nil
}

func (r *RazorpayClient) VerifyPayment(razorpayOrderID, razorpayPaymentID, razorpaySignature string) bool {

	data := razorpayOrderID + "|" + razorpayPaymentID

	h := hmac.New(sha256.New, []byte(r.KeySecret))
	h.Write([]byte(data))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(expectedSignature), []byte(razorpaySignature))
}

func (r *RazorpayClient) GetKeyID() string {
	return r.KeyID
}
