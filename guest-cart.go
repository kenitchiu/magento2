package magento2

import (
	"fmt"
	"github.com/hermsi1337/go-magento2/types"
	"strconv"
)

type GuestCart struct {
	QuoteID   string
	Detailed  types.DetailedCart
	ApiClient *ApiClient
}

func (cart *GuestCart) GetDetails() (types.DetailedCart, error) {
	var detailedCart = &types.DetailedCart{}
	httpClient := cart.ApiClient.httpClient
	resp, err := httpClient.R().SetResult(detailedCart).Get(guestCarts + "/" + cart.QuoteID)
	if err != nil {
		return *detailedCart, fmt.Errorf("error while getting detailed cart object from magento2-api: '%v'", err)
	} else if resp.StatusCode() >= 400 {
		return *detailedCart, fmt.Errorf("unexpected statuscode '%v' - response: '%v'", resp.StatusCode(), resp)
	}
	detailedCart = resp.Result().(*types.DetailedCart)

	return *detailedCart, err
}

func (cart *GuestCart) UpdateSelf() error {
	newDetails, err := cart.GetDetails()
	if err != nil {
		return fmt.Errorf("error while updating the cart object from magento2-api: '%v'", err)
	}

	cart.Detailed = newDetails
	return nil
}

func (cart *GuestCart) AddItems(items []types.Item) error {
	endpoint := guestCarts + "/" + cart.QuoteID + guestCartsItems
	httpClient := cart.ApiClient.httpClient

	type PayLoad struct {
		CartItem types.Item `json:"cartItem"`
	}

	for _, item := range items {
		item.QuoteID = cart.QuoteID
		payLoad := &PayLoad{
			CartItem: item,
		}
		resp, err := httpClient.R().SetBody(payLoad).Post(endpoint)
		if err != nil {
			return fmt.Errorf("received error while adding item '%v' to cart: '%v'", item, err)
		} else if resp.StatusCode() >= 400 {
			return fmt.Errorf("unexpected statuscode '%v' - response: '%v'", resp.StatusCode(), resp)
		}
	}

	err := cart.UpdateSelf()
	if err != nil {
		return err
	}

	return nil
}

func (cart *GuestCart) EstimateShippingCarrier(addrInfo types.Address) ([]types.Carrier, error) {
	endpoint := guestCarts + "/" + cart.QuoteID + guestCartsShippingCosts
	httpClient := cart.ApiClient.httpClient

	type PayLoad struct {
		Address types.Address `json:"address"`
	}

	payLoad := &PayLoad{
		Address: addrInfo,
	}

	shippingCosts := &[]types.Carrier{}

	resp, err := httpClient.R().SetBody(*payLoad).SetResult(shippingCosts).Post(endpoint)
	if err != nil {
		return *shippingCosts, fmt.Errorf("received erro while estimating shipping costs: '%v'", err)
	} else if resp.StatusCode() >= 400 {
		return *shippingCosts, fmt.Errorf("unexpected statuscode '%v' - response: '%v'", resp.StatusCode(), resp)
	}

	shippingCosts = resp.Result().(*[]types.Carrier)

	if len(*shippingCosts) == 0 {
		return *shippingCosts, fmt.Errorf("received no suitable shipping - response: '%v'", resp)
	}

	return *shippingCosts, nil
}

func (cart *GuestCart) AddShippingInformation(addrInfo types.AddressInformation) error {
	endpoint := guestCarts + "/" + cart.QuoteID + guestCartsShippingInformation
	httpClient := cart.ApiClient.httpClient

	type PayLoad struct {
		AddressInformation types.AddressInformation `json:"addressInformation"`
	}

	payLoad := &PayLoad{
		AddressInformation: addrInfo,
	}

	resp, err := httpClient.R().SetBody(*payLoad).Post(endpoint)
	if err != nil {
		return fmt.Errorf("received error while adding shipping information: '%v'", err)
	} else if resp.StatusCode() >= 400 {
		return fmt.Errorf("unexpected statuscode '%v' - response: '%v'", resp.StatusCode(), resp)
	}

	err = cart.UpdateSelf()
	if err != nil {
		return err
	}

	return nil
}

func (cart *GuestCart) EstimatePaymentMethods() ([]types.PaymentMethod, error) {
	endpoint := guestCarts + "/" + cart.QuoteID + guestCartsPaymentMethods
	httpClient := cart.ApiClient.httpClient

	paymentMethods := &[]types.PaymentMethod{}

	resp, err := httpClient.R().SetResult(paymentMethods).Get(endpoint)
	if err != nil {
		return *paymentMethods, fmt.Errorf("received error while estimating payment methods costs: '%v'", err)
	} else if resp.StatusCode() >= 400 {
		return *paymentMethods, fmt.Errorf("unexpected statuscode '%v' - response: '%v'", resp.StatusCode(), resp)
	}

	paymentMethods = resp.Result().(*[]types.PaymentMethod)

	if len(*paymentMethods) == 0 {
		return *paymentMethods, fmt.Errorf("received no suitable payment method - response: '%v'", resp)
	}

	return *paymentMethods, nil
}

func (cart *GuestCart) CreateOrder(paymentMethod types.PaymentMethod) (types.OrderID, error) {
	endpoint := guestCarts + "/" + cart.QuoteID + guestCartsOrder
	httpClient := cart.ApiClient.httpClient

	type PayLoad struct {
		PaymentMethod types.PaymentMethodCode `json:"paymentMethod"`
	}

	payLoad := &PayLoad{
		PaymentMethod: types.PaymentMethodCode{
			Method: paymentMethod.Code,
		},
	}

	resp, err := httpClient.R().SetBody(payLoad).Put(endpoint)
	if err != nil {
		return 0, fmt.Errorf("received error while creating order: '%v'", err)
	} else if resp.StatusCode() >= 400 {
		return 0, fmt.Errorf("unexpected statuscode '%v' - response: '%v'", resp.StatusCode(), resp)
	}

	orderIDString := mayTrimSurroundingQuotes(resp.String())
	orderIDInt, err := strconv.Atoi(orderIDString)
	if err != nil {
		return 0, fmt.Errorf("unexpected error while extracting orderID: '%v'", err)
	}
	orderID := types.OrderID(orderIDInt)

	return orderID, nil
}
