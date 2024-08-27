package providers

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"payment-service/app"
	"payment-service/domain/entities"
	"payment-service/domain/types"
	"payment-service/errors"
)

type AuthorizeNetPaymentProvider struct {
	endpoint string
}

type CreateTransactionRequest struct {
	XMLName                xml.Name                   `xml:"createTransactionRequest"`
	Xmlns                  string                     `xml:"xmlns,attr"`
	MerchantAuthentication MerchantAuthenticationType `xml:"merchantAuthentication"`
	TransactionRequest     TransactionRequestType     `xml:"transactionRequest"`
}

type MerchantAuthenticationType struct {
	Name           string `xml:"name"`
	TransactionKey string `xml:"transactionKey"`
}

type CreditCardType struct {
	CardNumber     string `xml:"cardNumber"`
	ExpirationDate string `xml:"expirationDate"`
	CardCode       string `xml:"cardCode"`
}

type PaymentType struct {
	CreditCard CreditCardType `xml:"creditCard"`
}

type TransactionRequestType struct {
	TransactionType string      `xml:"transactionType"`
	Amount          float64     `xml:"amount"`
	Payment         PaymentType `xml:"payment"`
}

type CreateTransactionResponse struct {
	XMLName             xml.Name             `xml:"createTransactionResponse"`
	Messages            Messages             `xml:"messages"`
	TransactionResponse *TransactionResponse `xml:"transactionResponse"`
}

type Messages struct {
	ResultCode string    `xml:"resultCode"`
	Message    []Message `xml:"message"`
}

type Message struct {
	Code string `xml:"code"`
	Text string `xml:"text"`
}

type TransactionResponse struct {
	ResponseCode   string    `xml:"responseCode"`
	AuthCode       string    `xml:"authCode"`
	AVSResultCode  string    `xml:"avsResultCode"`
	CVVResultCode  string    `xml:"cvvResultCode"`
	CAVVResultCode string    `xml:"cavvResultCode"`
	TransId        string    `xml:"transId"`
	AccountNumber  string    `xml:"accountNumber"`
	AccountType    string    `xml:"accountType"`
	Messages       []Message `xml:"messages>message"`
	NetworkTransId string    `xml:"networkTransId"`
}

func NewAuthorizeNetPaymentProvider() *AuthorizeNetPaymentProvider {
	return &AuthorizeNetPaymentProvider{
		endpoint: "https://apitest.authorize.net/xml/v1/request.api",
	}
}

func (self *AuthorizeNetPaymentProvider) Charge(params types.DepositParams, transaction entities.Transaction) (entities.Transaction, error) {
	request := CreateTransactionRequest{
		Xmlns: "AnetApi/xml/v1/schema/AnetApiSchema.xsd",
		MerchantAuthentication: MerchantAuthenticationType{
			Name:           app.App().Config().GetString("payment.authorize_login_id"),
			TransactionKey: app.App().Config().GetString("payment.authorize_transaction_key"),
		},
		TransactionRequest: TransactionRequestType{
			TransactionType: "authCaptureTransaction",
			Amount:          params.Amount,
			Payment: PaymentType{
				CreditCard: CreditCardType{
					CardNumber:     params.CreditCardNumber,
					ExpirationDate: params.ExpirationDate,
					CardCode:       params.CVV,
				},
			},
		},
	}

	maskedRequest := request
	maskedRequest.TransactionRequest.Payment.CreditCard.CardNumber = "****"
	maskedRequest.TransactionRequest.Payment.CreditCard.CardCode = "****"
	maskedRequest.TransactionRequest.Payment.CreditCard.ExpirationDate = "****"
	maskedRequestXml, _ := xml.MarshalIndent(maskedRequest, "", "    ")
	maskedRequestStr := string(maskedRequestXml)
	transaction.RequestPayload = &maskedRequestStr

	// Convert the request to XML
	requestXml, err := xml.MarshalIndent(request, "", "    ")
	if err != nil {
		app.App().Logger().Error("failed to marshal XML: ", err.Error())
		return transaction, &errors.ValidationError{
			Message: "failed to marshal XML: " + err.Error(),
		}
	}

	xmlHeader := []byte(xml.Header)
	fullRequestXml := append(xmlHeader, requestXml...)

	httpClient := &http.Client{}
	req, err := http.NewRequest("POST", self.endpoint, bytes.NewBuffer(fullRequestXml))
	if err != nil {
		app.App().Logger().Error("failed to create HTTP request: ", err.Error())
		return transaction, &errors.ValidationError{
			Message: "failed to create HTTP request: " + err.Error(),
		}
	}
	req.Header.Set("Content-Type", "text/xml")

	resp, err := httpClient.Do(req)
	if err != nil {
		app.App().Logger().Error("failed to send HTTP request: ", err.Error())
		return transaction, &errors.ValidationError{
			Message: "failed to send HTTP request: " + err.Error(),
		}
	}
	defer resp.Body.Close()

	responseXml, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		app.App().Logger().Error("failed to read HTTP response: ", err.Error())
		return transaction, &errors.ValidationError{
			Message: "failed to read HTTP response: " + err.Error(),
		}
	}

	response := new(CreateTransactionResponse)
	err = xml.Unmarshal(responseXml, response)
	if err != nil {
		app.App().Logger().Error("failed to unmarshal XML response: ", err.Error())
		return transaction, &errors.ValidationError{
			Message: "failed to unmarshal XML response: " + err.Error(),
		}
	}

	if response.TransactionResponse != nil {
		transaction.PaymentId = response.TransactionResponse.TransId
	} else {
		app.App().Logger().Error("transaction failed: no transaction ID returned")
		return transaction, &errors.ValidationError{
			Message: "transaction failed: no transaction ID returned",
		}
	}

	responseStr := string(responseXml)
	transaction.ResponsePayload = &responseStr

	if response.TransactionResponse.ResponseCode == "1" {
		transaction.PaymentId = response.TransactionResponse.TransId
		transaction.Status = "succeeded"
	} else {
		transaction.Status = "failed"
	}

	return transaction, nil
}

func (self *AuthorizeNetPaymentProvider) Withdraw(params types.WithdrawParams, transaction entities.Transaction) (entities.Transaction, error) {
	request := CreateTransactionRequest{
		Xmlns: "AnetApi/xml/v1/schema/AnetApiSchema.xsd",
		MerchantAuthentication: MerchantAuthenticationType{
			Name:           app.App().Config().GetString("payment.authorize_login_id"),
			TransactionKey: app.App().Config().GetString("payment.authorize_transaction_key"),
		},
		TransactionRequest: TransactionRequestType{
			TransactionType: "refundTransaction",
			Amount:          float64(params.Amount),
			Payment: PaymentType{
				CreditCard: CreditCardType{
					CardNumber:     params.CreditCardNumber,
					ExpirationDate: params.ExpirationDate,
					CardCode:       params.CVV,
				},
			},
		},
	}

	maskedRequest := request
	maskedRequest.TransactionRequest.Payment.CreditCard.CardNumber = "****"
	maskedRequest.TransactionRequest.Payment.CreditCard.CardCode = "****"
	maskedRequest.TransactionRequest.Payment.CreditCard.ExpirationDate = "****"
	maskedRequestXml, _ := xml.MarshalIndent(maskedRequest, "", "    ")
	maskedRequestStr := string(maskedRequestXml)
	transaction.RequestPayload = &maskedRequestStr

	requestXml, err := xml.MarshalIndent(request, "", "    ")
	if err != nil {
		app.App().Logger().Error("failed to marshal XML: ", err.Error())
		return transaction, &errors.ValidationError{
			Message: "failed to marshal XML: " + err.Error(),
		}
	}

	xmlHeader := []byte(xml.Header)
	fullRequestXml := append(xmlHeader, requestXml...)

	httpClient := &http.Client{}
	req, err := http.NewRequest("POST", self.endpoint, bytes.NewBuffer(fullRequestXml))
	if err != nil {
		app.App().Logger().Error("failed to create HTTP request: ", err.Error())
		return transaction, &errors.ValidationError{
			Message: "failed to create HTTP request: " + err.Error(),
		}
	}
	req.Header.Set("Content-Type", "text/xml")

	resp, err := httpClient.Do(req)
	if err != nil {
		app.App().Logger().Error("failed to send HTTP request: ", err.Error())
		return transaction, &errors.ValidationError{
			Message: "failed to send HTTP request: " + err.Error(),
		}
	}
	defer resp.Body.Close()

	responseXml, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		app.App().Logger().Error("failed to read HTTP response: ", err.Error())
		return transaction, &errors.ValidationError{
			Message: "failed to read HTTP response: " + err.Error(),
		}
	}

	response := new(CreateTransactionResponse)
	err = xml.Unmarshal(responseXml, response)
	if err != nil {
		app.App().Logger().Error("failed to unmarshal XML response: ", err.Error())
		return transaction, &errors.ValidationError{
			Message: "failed to unmarshal XML response: " + err.Error(),
		}
	}

	if response.TransactionResponse != nil {
		transaction.PaymentId = response.TransactionResponse.TransId
	} else {
		app.App().Logger().Error("transaction failed: no transaction ID returned")
		return transaction, &errors.ValidationError{
			Message: "transaction failed: no transaction ID returned",
		}
	}

	responseStr := string(responseXml)
	transaction.ResponsePayload = &responseStr

	if response.TransactionResponse.ResponseCode == "1" {
		transaction.PaymentId = response.TransactionResponse.TransId
		transaction.Status = "succeeded"
	} else {
		transaction.Status = "failed"
	}

	return transaction, nil
}
