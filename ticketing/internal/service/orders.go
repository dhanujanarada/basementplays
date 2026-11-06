package service

import (
	"context"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/dhanujanarada/basementplays/ticketing/internal/domain"
	"github.com/dhanujanarada/basementplays/ticketing/internal/store"
)

const (
	maxBuyerNameLength  = 120
	maxBuyerPhoneLength = 32
	maxBuyerEmailLength = 254
	maxNotesLength      = 2000
	maxCancelReason     = 500
)

var (
	emailPattern = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

	publicPaymentMethods = map[string]bool{
		"payhere":       true,
		"bank_transfer": true,
	}

	manualPaymentMethods = map[string]bool{
		"bank_transfer": true,
		"cash":          true,
		"whatsapp":      true,
	}

	allPaymentMethods = map[string]bool{
		"payhere":       true,
		"bank_transfer": true,
		"cash":          true,
		"whatsapp":      true,
	}
)

// PublicTicket is the buyer-visible view of an issued ticket. QR tokens are
// added in step 6.
type PublicTicket struct {
	ID       string  `json:"ticket_id"`
	SeatCode string  `json:"seat_code"`
	Status   string  `json:"status"`
	IssuedAt string  `json:"issued_at"`
	UsedAt   *string `json:"used_at"`
}

// PublicOrder is the buyer-visible order shape. It deliberately omits buyer
// contact details and internal notes.
type PublicOrder struct {
	ID         string            `json:"order_id"`
	ShowID     string            `json:"show_id"`
	Status     string            `json:"status"`
	Source     string            `json:"source"`
	TotalCents int               `json:"total_cents"`
	Currency   string            `json:"currency"`
	ExpiresAt  string            `json:"expires_at"`
	PaidAt     *string           `json:"paid_at"`
	CreatedAt  string            `json:"created_at"`
	Seats      []store.OrderSeat `json:"seats"`
	Tickets    []PublicTicket    `json:"tickets"`
}

// InternalOrder is the full order shape for Backstage pages.
type InternalOrder struct {
	ID            string            `json:"order_id"`
	ShowID        string            `json:"show_id"`
	Status        string            `json:"status"`
	Source        string            `json:"source"`
	PaymentMethod *string           `json:"payment_method"`
	BuyerName     *string           `json:"buyer_name"`
	BuyerPhone    *string           `json:"buyer_phone"`
	BuyerEmail    *string           `json:"buyer_email"`
	TotalCents    int               `json:"total_cents"`
	Currency      string            `json:"currency"`
	ExpiresAt     string            `json:"expires_at"`
	PaidAt        *string           `json:"paid_at"`
	Notes         string            `json:"notes"`
	CreatedAt     string            `json:"created_at"`
	Seats         []store.OrderSeat `json:"seats"`
	Tickets       []InternalTicket  `json:"tickets"`
}

// InternalTicket is a ticket row as shown in Backstage.
type InternalTicket struct {
	ID       string  `json:"ticket_id"`
	SeatCode string  `json:"seat_code"`
	Status   string  `json:"status"`
	IssuedAt string  `json:"issued_at"`
	UsedAt   *string `json:"used_at"`
}

// CheckoutInput captures buyer details at checkout.
type CheckoutInput struct {
	BuyerName     string `json:"buyer_name"`
	BuyerPhone    string `json:"buyer_phone"`
	BuyerEmail    string `json:"buyer_email"`
	PaymentMethod string `json:"payment_method"`
}

// ManualOrderInput is the internal manual-order payload.
type ManualOrderInput struct {
	ShowID        string   `json:"show_id"`
	Seats         []string `json:"seats"`
	BuyerName     string   `json:"buyer_name"`
	BuyerPhone    string   `json:"buyer_phone"`
	BuyerEmail    string   `json:"buyer_email"`
	PaymentMethod string   `json:"payment_method"`
	Notes         string   `json:"notes"`
}

// Orders implements order, checkout, payment, and cancellation use cases.
type Orders struct {
	store               *store.Store
	checkoutHoldMinutes int
	manualHoldHours     int
	maxSeats            int
}

// NewOrders builds the order service.
func NewOrders(st *store.Store, checkoutHoldMinutes, manualHoldHours, maxSeats int) *Orders {
	return &Orders{
		store:               st,
		checkoutHoldMinutes: checkoutHoldMinutes,
		manualHoldHours:     manualHoldHours,
		maxSeats:            maxSeats,
	}
}

// Checkout records buyer details and extends a pending hold. A PayHere checkout
// keeps the online source and uses CHECKOUT_HOLD_MINUTES; bank transfer becomes
// a manual order and uses MANUAL_HOLD_HOURS.
func (o *Orders) Checkout(
	ctx context.Context,
	orderID, orderToken string,
	input CheckoutInput,
	idempotencyKey string,
) (PublicOrder, error) {
	orderID = strings.TrimSpace(orderID)
	if orderID == "" {
		return PublicOrder{}, domain.Invalid("order_id", "is required")
	}
	if len(orderID) > maxOrderIDLength {
		return PublicOrder{}, domain.Invalid("order_id", "is invalid")
	}

	orderToken = strings.TrimSpace(orderToken)
	if orderToken == "" {
		return PublicOrder{}, domain.Invalid("order_token", "is required")
	}

	input = normalizeCheckoutInput(input)
	if err := validateBuyer(input.BuyerName, input.BuyerPhone, input.BuyerEmail); err != nil {
		return PublicOrder{}, err
	}
	if !publicPaymentMethods[input.PaymentMethod] {
		return PublicOrder{}, domain.Invalid("payment_method", `must be one of "payhere" or "bank_transfer"`)
	}

	key, err := normalizeIdempotencyKey(idempotencyKey)
	if err != nil {
		return PublicOrder{}, err
	}

	source := "online"
	expiresAt := time.Now().Add(time.Duration(o.checkoutHoldMinutes) * time.Minute)
	if input.PaymentMethod == "bank_transfer" {
		source = "manual"
		expiresAt = time.Now().Add(time.Duration(o.manualHoldHours) * time.Hour)
	}

	order, err := o.store.Checkout(ctx, store.CheckoutInput{
		OrderID:        orderID,
		TokenHash:      hashSecretToken(orderToken),
		BuyerName:      input.BuyerName,
		BuyerPhone:     input.BuyerPhone,
		BuyerEmail:     input.BuyerEmail,
		PaymentMethod:  input.PaymentMethod,
		Source:         source,
		ExpiresAt:      expiresAt,
		IdempotencyKey: key,
	})
	if err != nil {
		return PublicOrder{}, err
	}

	return toPublicOrder(order), nil
}

// Get returns an order after checking the buyer's order token.
func (o *Orders) Get(ctx context.Context, orderID, orderToken string) (PublicOrder, error) {
	orderID = strings.TrimSpace(orderID)
	if orderID == "" {
		return PublicOrder{}, domain.Invalid("order_id", "is required")
	}
	if len(orderID) > maxOrderIDLength {
		return PublicOrder{}, domain.Invalid("order_id", "is invalid")
	}

	orderToken = strings.TrimSpace(orderToken)
	if orderToken == "" {
		return PublicOrder{}, domain.Invalid("order_token", "is required")
	}

	order, err := o.store.GetOrderForToken(ctx, orderID, hashSecretToken(orderToken))
	if err != nil {
		return PublicOrder{}, err
	}

	return toPublicOrder(order), nil
}

// CreateManual creates an order taken by staff. The generated order token is
// discarded; staff use the internal API instead.
func (o *Orders) CreateManual(ctx context.Context, input ManualOrderInput) (InternalOrder, error) {
	input.ShowID = strings.TrimSpace(input.ShowID)
	if input.ShowID == "" {
		return InternalOrder{}, domain.Invalid("show_id", "is required")
	}
	if len(input.ShowID) > maxIDLength {
		return InternalOrder{}, domain.Invalid("show_id", "must be at most %d characters", maxIDLength)
	}

	seats, err := normalizeSeats(input.Seats, o.maxSeats)
	if err != nil {
		return InternalOrder{}, err
	}

	input.BuyerName = strings.TrimSpace(input.BuyerName)
	input.BuyerPhone = strings.TrimSpace(input.BuyerPhone)
	input.BuyerEmail = strings.TrimSpace(input.BuyerEmail)
	input.PaymentMethod = strings.TrimSpace(input.PaymentMethod)
	input.Notes = strings.TrimSpace(input.Notes)

	if err := validateBuyer(input.BuyerName, input.BuyerPhone, input.BuyerEmail); err != nil {
		return InternalOrder{}, err
	}
	if !manualPaymentMethods[input.PaymentMethod] {
		return InternalOrder{}, domain.Invalid(
			"payment_method",
			`must be one of "bank_transfer", "cash", or "whatsapp"`,
		)
	}
	if utf8.RuneCountInString(input.Notes) > maxNotesLength {
		return InternalOrder{}, domain.Invalid("notes", "must be at most %d characters", maxNotesLength)
	}

	token, err := newSecretToken()
	if err != nil {
		return InternalOrder{}, err
	}

	order, _, err := o.store.CreateOrder(ctx, store.CreateOrderInput{
		ShowID:          input.ShowID,
		SeatCodes:       seats,
		Source:          "manual",
		PaymentMethod:   &input.PaymentMethod,
		BuyerName:       &input.BuyerName,
		BuyerPhone:      &input.BuyerPhone,
		BuyerEmail:      stringPtrOrNil(input.BuyerEmail),
		Notes:           input.Notes,
		ExpiresAt:       time.Now().Add(time.Duration(o.manualHoldHours) * time.Hour),
		AccessTokenHash: hashSecretToken(token),
	})
	if err != nil {
		return InternalOrder{}, err
	}

	return toInternalOrder(order), nil
}

// MarkPaid sells the order's seats and issues tickets. It is idempotent; a
// payment whose seats were taken returns an order in needs_review.
func (o *Orders) MarkPaid(ctx context.Context, orderID string, paymentMethod *string) (InternalOrder, error) {
	orderID = strings.TrimSpace(orderID)
	if orderID == "" {
		return InternalOrder{}, domain.Invalid("order_id", "is required")
	}
	if len(orderID) > maxOrderIDLength {
		return InternalOrder{}, domain.Invalid("order_id", "is invalid")
	}

	if paymentMethod != nil {
		method := strings.TrimSpace(*paymentMethod)
		if method == "" {
			paymentMethod = nil
		} else {
			if !allPaymentMethods[method] {
				return InternalOrder{}, domain.Invalid(
					"payment_method",
					`must be one of "payhere", "bank_transfer", "cash", or "whatsapp"`,
				)
			}
			paymentMethod = &method
		}
	}

	result, err := o.store.MarkPaid(ctx, store.MarkPaidInput{
		OrderID:       orderID,
		PaymentMethod: paymentMethod,
	})
	if err != nil {
		return InternalOrder{}, err
	}

	return toInternalOrder(result.Order), nil
}

// Cancel cancels an order and releases or voids its seats and tickets.
func (o *Orders) Cancel(ctx context.Context, orderID, reason string) (InternalOrder, error) {
	orderID = strings.TrimSpace(orderID)
	if orderID == "" {
		return InternalOrder{}, domain.Invalid("order_id", "is required")
	}
	if len(orderID) > maxOrderIDLength {
		return InternalOrder{}, domain.Invalid("order_id", "is invalid")
	}

	reason = strings.TrimSpace(reason)
	if utf8.RuneCountInString(reason) > maxCancelReason {
		return InternalOrder{}, domain.Invalid("reason", "must be at most %d characters", maxCancelReason)
	}

	order, err := o.store.CancelOrder(ctx, orderID, reason)
	if err != nil {
		return InternalOrder{}, err
	}

	return toInternalOrder(order), nil
}

// List returns every order for a show, newest first.
func (o *Orders) List(ctx context.Context, showID string) ([]InternalOrder, error) {
	showID = strings.TrimSpace(showID)
	if showID == "" {
		return nil, domain.Invalid("show_id", "is required")
	}
	if len(showID) > maxIDLength {
		return nil, domain.Invalid("show_id", "must be at most %d characters", maxIDLength)
	}

	orders, err := o.store.ListShowOrders(ctx, showID)
	if err != nil {
		return nil, err
	}

	out := make([]InternalOrder, 0, len(orders))
	for _, order := range orders {
		out = append(out, toInternalOrder(order))
	}

	return out, nil
}

func normalizeCheckoutInput(input CheckoutInput) CheckoutInput {
	input.BuyerName = strings.TrimSpace(input.BuyerName)
	input.BuyerPhone = strings.TrimSpace(input.BuyerPhone)
	input.BuyerEmail = strings.TrimSpace(input.BuyerEmail)
	input.PaymentMethod = strings.TrimSpace(input.PaymentMethod)
	return input
}

func validateBuyer(name, phone, email string) error {
	name = strings.TrimSpace(name)
	phone = strings.TrimSpace(phone)
	email = strings.TrimSpace(email)

	if name == "" {
		return domain.Invalid("buyer_name", "is required")
	}
	if utf8.RuneCountInString(name) > maxBuyerNameLength {
		return domain.Invalid("buyer_name", "must be at most %d characters", maxBuyerNameLength)
	}
	if hasControl(name) {
		return domain.Invalid("buyer_name", "must not contain control characters")
	}

	if phone == "" {
		return domain.Invalid("buyer_phone", "is required")
	}
	if utf8.RuneCountInString(phone) > maxBuyerPhoneLength {
		return domain.Invalid("buyer_phone", "must be at most %d characters", maxBuyerPhoneLength)
	}
	if hasControl(phone) {
		return domain.Invalid("buyer_phone", "must not contain control characters")
	}

	if email == "" {
		return nil
	}
	if len(email) > maxBuyerEmailLength {
		return domain.Invalid("buyer_email", "must be at most %d characters", maxBuyerEmailLength)
	}
	if !emailPattern.MatchString(email) {
		return domain.Invalid("buyer_email", "must be a valid email address")
	}

	return nil
}

func hasControl(value string) bool {
	for _, r := range value {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}

func stringPtrOrNil(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func formatTime(value time.Time) string {
	return value.Format(timeFormat)
}

func formatOptionalTime(value *time.Time) *string {
	if value == nil {
		return nil
	}

	formatted := value.Format(timeFormat)
	return &formatted
}

func toPublicOrder(order store.Order) PublicOrder {
	tickets := make([]PublicTicket, 0, len(order.Tickets))
	for _, ticket := range order.Tickets {
		tickets = append(tickets, PublicTicket{
			ID:       ticket.ID,
			SeatCode: ticket.SeatCode,
			Status:   ticket.Status,
			IssuedAt: formatTime(ticket.IssuedAt),
			UsedAt:   formatOptionalTime(ticket.UsedAt),
		})
	}

	return PublicOrder{
		ID:         order.ID,
		ShowID:     order.ShowID,
		Status:     order.Status,
		Source:     order.Source,
		TotalCents: order.TotalCents,
		Currency:   order.Currency,
		ExpiresAt:  formatTime(order.ExpiresAt),
		PaidAt:     formatOptionalTime(order.PaidAt),
		CreatedAt:  formatTime(order.CreatedAt),
		Seats:      order.Seats,
		Tickets:    tickets,
	}
}

func toInternalOrder(order store.Order) InternalOrder {
	tickets := make([]InternalTicket, 0, len(order.Tickets))
	for _, ticket := range order.Tickets {
		tickets = append(tickets, InternalTicket{
			ID:       ticket.ID,
			SeatCode: ticket.SeatCode,
			Status:   ticket.Status,
			IssuedAt: formatTime(ticket.IssuedAt),
			UsedAt:   formatOptionalTime(ticket.UsedAt),
		})
	}

	return InternalOrder{
		ID:            order.ID,
		ShowID:        order.ShowID,
		Status:        order.Status,
		Source:        order.Source,
		PaymentMethod: order.PaymentMethod,
		BuyerName:     order.BuyerName,
		BuyerPhone:    order.BuyerPhone,
		BuyerEmail:    order.BuyerEmail,
		TotalCents:    order.TotalCents,
		Currency:      order.Currency,
		ExpiresAt:     formatTime(order.ExpiresAt),
		PaidAt:        formatOptionalTime(order.PaidAt),
		Notes:         order.Notes,
		CreatedAt:     formatTime(order.CreatedAt),
		Seats:         order.Seats,
		Tickets:       tickets,
	}
}
