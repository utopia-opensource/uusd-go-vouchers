package uusdv

import (
	"errors"
	"math"

	utopiago "github.com/Sagleft/utopialib-go"
)

// VoucherData contains voucher data ^ↀᴥↀ^
type VoucherData struct {
	Status           string  `json:"status"` //"pending" or "done"
	CreatedTimestamp string  `json:"created"`
	Amount           float64 `json:"amount"`
	Comments         string  `json:"comments"`
	Direction        int     `json:"direction"`
	TransactionID    string  `json:"trid"`
}

// ActivationData contains voucher activation data (=ↀωↀ=)
type ActivationData struct {
	Status          string  `json:"status"` //"pending" or "done"
	ReferenceNumber string  `json:"referenceNumber"`
	Amount          float64 `json:"amount"`
}

// Handler is a handler for all requests for UUSD voucher management
type Handler struct {
	Client *utopiago.UtopiaClient
}

// SetClient - connects another Utopia client to Handler
func (h *Handler) SetClient(client *utopiago.UtopiaClient) error {
	if !client.CheckClientConnection() {
		return errors.New("client disconnected")
	}
	h.Client = client
	return nil
}

// ActivateVoucher - an attempt to activate a voucher and get data about it
func (h *Handler) ActivateVoucher(voucherCode string) (ActivationData, error) {
	referenceNumber, err := h.Client.UseVoucher(voucherCode)
	if err != nil {
		return ActivationData{}, err
	}
	return h.CheckVoucherActivation(referenceNumber)
}

// CheckVoucherActivation - lite version of CheckVoucherStatus
func (h *Handler) CheckVoucherActivation(referenceNumber string) (ActivationData, error) {
	data, err := h.CheckVoucherStatus(referenceNumber)
	if err != nil {
		return ActivationData{}, err
	}
	//TODO: add fields exists check
	return ActivationData{
		Status:          data.Status,
		ReferenceNumber: referenceNumber,
		Amount:          data.Amount,
	}, nil
}

func (h *Handler) getVoucherDataMap(referenceNumber string) (map[string]interface{}, error) {
	voucherDataRaw, err := h.Client.GetFinanceHistory("ALL_VOUCHERS", referenceNumber)
	if err != nil {
		return nil, err
	}
	if len(voucherDataRaw) == 0 {
		return nil, errors.New("finance history not found")
	}
	firstElement := voucherDataRaw[0]

	voucherDataMap, ok := firstElement.(map[string]interface{})
	if !ok {
		return nil, errors.New("can't find voucher data in client response")
	}
	return voucherDataMap, nil
}

// CheckVoucherStatus - checks the voucher data, its activation status
func (h *Handler) CheckVoucherStatus(referenceNumber string) (VoucherData, error) {
	voucherDataMap, err := h.getVoucherDataMap(referenceNumber)
	if err != nil {
		return VoucherData{}, err
	}
	//TODO: check fields exists
	resultData := VoucherData{
		CreatedTimestamp: voucherDataMap["created"].(string),
		Amount:           voucherDataMap["amount"].(float64),
		Comments:         voucherDataMap["comments"].(string),
		Direction:        int(math.Round(voucherDataMap["direction"].(float64))),
		TransactionID:    voucherDataMap["id"].(string),
	}
	if voucherDataMap["state"] == "-1" || voucherDataMap["state"] == -1 {
		resultData.Status = "pending"
	}
	if voucherDataMap["state"] == "0" || voucherDataMap["state"] == 0 {
		resultData.Status = "done"
	}

	return resultData, nil
}

// GetVoucherAmount - asks for the amount of the voucher if it has already been activated
func (h *Handler) GetVoucherAmount(referenceNumber string) (float64, error) {
	vData, err := h.CheckVoucherStatus(referenceNumber)
	if err != nil {
		return 0, err
	}
	return vData.Amount, nil
}

// CreateVoucher - an attempt to create a voucher for a given amount
func (h *Handler) CreateVoucher(amount float64) (ActivationData, error) {
	accountBalance, err := h.Client.GetBalance()
	if err != nil {
		return ActivationData{}, err
	}
	if accountBalance < amount {
		return ActivationData{}, errors.New("not enough balance")
	}
	referenceNumber, err := h.Client.CreateUUSDVoucher(amount)
	if err != nil {
		return ActivationData{}, err
	}
	return h.CheckVoucherActivation(referenceNumber)
}

// GetNetFee - asks for a commission to create a voucher with a specific amount
func (h *Handler) GetNetFee(amount float64) (float64, error) {
	//TODO: user data from client
	return amount * 0.0015, nil
}
