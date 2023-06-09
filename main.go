package main

import (
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/manifoldco/promptui"
)

type Object struct {
	Invoice invoice `json:"invoice"`
}

type invoice struct {
	CorrelationID string `json:"@_documentCorrelationId" xml:"documentCorrelationId,attr"`
	DocTypeID     string `json:"@_docTypeId" xml:"docTypeId,attr"`
	DocInstanceID int64  `json:"@_docInstanceId" xml:"docInstanceId,attr"`
	DocPlatform   string `json:"@_docPlatform" xml:"docPlatform,attr"`
	Serie         string `json:"@_serie" xml:"serie,attr"`
	CurrencyISO   string `json:"currencyISOCode" xml:"currencyISOCode"`
	References    struct {
		ThirdPartyErpInternalReference string `json:"thirdPartyErpInternalReference"`
	} `json:"documentReferences" xml:"documentReferences"`
	Dates struct {
		DocumentDate              string `json:"documentDate" xml:"documentDate"`
		GoodsServiceAvailableDate string `json:"goodsServiceAvailableDate" xml:"goodsServiceAvailableDate"`
		DueDate                   string `json:"dueDate" xml:"dueDate"`
	} `json:"documentDates" xml:"documentDates"`
	Parties struct {
		Seller struct {
			Name                        string `json:"name" xml:"name"`
			Country                     string `json:"country" xml:"country"`
			VATNumber                   string `json:"vatNumber" xml:"vatNumber"`
			Address                     string `json:"address" xml:"address"`
			City                        string `json:"city" xml:"city"`
			ZipArea                     string `json:"zipArea" xml:"zipArea"`
			ZipCode                     string `json:"zipCode" xml:"zipCode"`
			CompanyRegistrationNumber   string `json:"companyRegistrationNumber" xml:"companyRegistrationNumber"`
			CompanyRegistrationLocation string `json:"companyRegistrationLocation" xml:"companyRegistrationLocation"`
			SocialCapitalValue          string `json:"socialCapitalValue" xml:"socialCapitalValue"`
		} `json:"seller" xml:"seller"`
		Buyer struct {
			Name      string `json:"name" xml:"name"`
			Email     string `json:"email" xml:"email"`
			Country   string `json:"country" xml:"country"`
			VATNumber string `json:"vatNumber" xml:"vatNumber"`
			Address   string `json:"address" xml:"address"`
			ZipArea   string `json:"zipArea" xml:"zipArea"`
			ZipCode   string `json:"zipCode" xml:"zipCode"`
		} `json:"buyer" xml:"buyer"`
	} `json:"partyInformation" xml:"partyInformation"`
	LineItems []struct {
		Number                    int64   `json:"@_number" xml:"number,attr"`
		SellerAssignedTradeItemID string  `json:"sellerAssignedTradeItemIdentification" xml:"sellerAssignedTradeItemIdentification"`
		ItemDescription           string  `json:"itemDescription" xml:"itemDescription"`
		NetPrice                  float64 `json:"netPrice" xml:"netPrice"`
		NetLineAmount             float64 `json:"netLineAmount" xml:"netLineAmount"`
		GrossPrice                float64 `json:"grossPrice" xml:"grossPrice"`
		GrossLineAmount           float64 `json:"grossLineAmount" xml:"grossLineAmount"`
		LineTotalPayableAmount    float64 `json:"lineTotalPayableAmount" xml:"lineTotalPayableAmount"`
		Quantity                  struct {
			Value         int    `json:"value" xml:"value"`
			UnitCodeValue string `json:"unitCodeValue" xml:"unitCodeValue"`
		} `json:"quantity" xml:"quantity"`
		LineVat struct {
			TaxableAmount float64 `json:"taxableAmount" xml:"taxableAmount"`
			TaxPercentage float64 `json:"taxPercentage" xml:"taxPercentage"`
			TaxTotalValue float64 `json:"taxTotalValue" xml:"taxTotalValue"`
		} `json:"lineVat" xml:"lineVat"`
	} `json:"lineItem" xml:"lineItem"`
	Totals struct {
		NumberOfLines         int64   `json:"numberOfLines" xml:"numberOfLines"`
		TotalAmountPayable    float64 `json:"totalAmountPayable" xml:"totalAmountPayable"`
		TotalVatTaxableAmount float64 `json:"totalVatTaxableAmount" xml:"totalVatTaxableAmount"`
		TotalVatAmount        float64 `json:"totalVatAmount" xml:"totalVatAmount"`
		TotalGrossAmount      float64 `json:"totalGrossAmount" xml:"totalGrossAmount"`
		TotalNetAmount        float64 `json:"totalNetAmount" xml:"totalNetAmount"`
		VatSummary            struct {
			TaxPercentage float64 `json:"taxPercentage" xml:"taxPercentage"`
			TaxTotalValue float64 `json:"taxTotalValue" xml:"taxTotalValue"`
			TaxableAmount float64 `json:"taxableAmount" xml:"taxableAmount"`
		} `json:"vatSummary" xml:"vatSummary"`
	} `json:"documentTotals" xml:"documentTotals"`
	EmailNotification struct {
		Email        string `json:"email" xml:"email"`
		LanguageCode string `json:"languageCode" xml:"languageCode"`
	} `json:"emailNotification" xml:"emailNotification"`
}

type MessageResponseObject struct {
	Result MessageBodyObject `json:"ResultData"`
}

type MessagesResponseObject struct {
	Results MessageHeaderObject `json:"ResultData"`
}

type MessageHeaderObject struct {
	Messages []MessageBodyObject `json:"MessageIds"`
}

type MessageBodyObject struct {
	ID         string `json:"MessageId"`
	Receiver   string `json:"Receiver"`
	Sender     string `json:"Sender"`
	Base64Data string `json:"Base64Data,omitempty"`
}

func main() {
	var payloadType string
	var jsonString string
	var getMessages bool
	var clearMessages bool
	var getMessageId string
	var processMessageId string

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	flag.StringVar(&payloadType, "send", "", "Type of message to send")
	flag.BoolVar(&getMessages, "getmessages", false, "Get all message IDs")
	flag.BoolVar(&clearMessages, "clearmessages", false, "Clear all messages in the queue")
	flag.StringVar(&getMessageId, "getmessage", "", "Get message by ID")
	flag.StringVar(&processMessageId, "processmessage", "", "Process message by ID")

	flag.Parse()

	if payloadType != "" {
		jsonString = flag.Arg(0)
		if jsonString == "" {
			log.Fatal("No JSON string provided")
		}
		sendMessage(payloadType, jsonString)
	} else if getMessages {
		messages := getAllMessages()
		messageJSON, err := json.MarshalIndent(messages, "", "  ")
		if err != nil {
			log.Fatalf(err.Error())
		}
		fmt.Println(string(messageJSON))
	} else if getMessageId != "" {
		decode := yesNo("Do you want to return just the Base64 decoded response?")
		message, decodedResult := getMessageById(getMessageId, decode)
		if decode {
			fmt.Println(decodedResult)
		} else {
			messageJSON, err := json.MarshalIndent(message, "", "  ")
			if err != nil {
				log.Fatalf(err.Error())
			}
			fmt.Println(string(messageJSON))
		}
	} else if processMessageId != "" {
		if !processMessageById(processMessageId) {
			fmt.Printf("Failed to clear message %s from the queue\n", processMessageId)
		} else {
			fmt.Printf("Message %s cleared from the queue\n", processMessageId)
		}
	} else if clearMessages {
		clearAllMessages()
	} else {
		fmt.Println("Invalid arguments")
		flag.Usage()
	}
}

func decodePayloadReseponse(base64String string) string {
	decoded, err := base64.StdEncoding.DecodeString(base64String)
	if err != nil {
		panic(err)
	}

	return string(decoded)
}

func sendMessage(payloadType string, jsonString string) {
	var xmlData []byte
	var err error
	var fileName string

	switch payloadType {
	case "invoice":
		payload := Object{}
		if err := json.Unmarshal([]byte(jsonString), &payload); err != nil {
			panic(err)
		}
		fileName = payload.Invoice.References.ThirdPartyErpInternalReference
		xmlData, err = xml.MarshalIndent(payload.Invoice, "", " ")
		if err != nil {
			panic(err)
		}
	default:
		fmt.Println("Unknown payload type")
		os.Exit(1)
	}

	xmlPayload := fmt.Sprintf("%s%s", "", xmlData)

	url := fmt.Sprintf(
		"%s%s",
		os.Getenv("SOVOS_BASE_URL"),
		os.Getenv("SOVOS_RECEIVE_ENDPOINT"),
	)
	method := "POST"
	id := uuid.New()
	postBody := strings.NewReader(fmt.Sprintf(
		`{
			"Sender": "%s",
			"Receiver": "%s",
			"ContentType": "application/xml",
			"Base64Data": "%s",
			"MessageId": "%s",
			"Filename": "%v.xml"
		}`,
		os.Getenv("SOVOS_USER_ID"),
		os.Getenv("SOVOS_ENVIRONMENT"),
		base64.StdEncoding.EncodeToString([]byte(xmlPayload)),
		id.String(),
		fileName,
	))

	response, err := makeRequest(method, url, postBody)
	if err != nil {
		log.Fatalf("Error with PROCESSQUEUEDMESSAGE call: %s", err)
	}

	fmt.Println(xmlPayload)
	fmt.Println(postBody)
	fmt.Println(string(response))
}

func getAllMessages() MessagesResponseObject {
	url := fmt.Sprintf(
		"%s%s?Receiver=%s",
		os.Getenv("SOVOS_BASE_URL"),
		os.Getenv("SOVOS_GET_MESSAGES_ENDPOINT"),
		os.Getenv("SOVOS_USER_ID"),
	)
	method := "GET"
	body, err := makeRequest(method, url, nil)
	if err != nil {
		log.Fatalf("Error with GETQUEUEDMESSAGES call: %s", err)
	}

	messages := MessagesResponseObject{}
	if err := json.Unmarshal(body, &messages); err != nil {
		panic(err)
	}

	return messages
}

func getMessageById(messageID string, decode bool) (MessageBodyObject, string) {
	for _, result := range getAllMessages().Results.Messages {
		if messageID == result.ID {
			url := fmt.Sprintf(
				"%s%s?Receiver=%s&Sender=%s&MessageID=%s",
				os.Getenv("SOVOS_BASE_URL"),
				os.Getenv("SOVOS_GET_MESSAGE_ENDPOINT"),
				result.Receiver,
				result.Sender,
				messageID,
			)
			method := "GET"
			body, err := makeRequest(method, url, nil)
			if err != nil {
				log.Fatalf("Error with GETMESSAGEDATA call: %s", err)
			}
			message := MessageResponseObject{}
			if err := json.Unmarshal(body, &message); err != nil {
				panic(err)
			}
			if decode {
				return MessageBodyObject{}, decodePayloadReseponse(message.Result.Base64Data)
			}

			return message.Result, ""
		}
	}

	return MessageBodyObject{}, ""
}

func processMessageById(messageID string) bool {
	for _, result := range getAllMessages().Results.Messages {
		if messageID == result.ID {
			url := fmt.Sprintf(
				"%s%s",
				os.Getenv("SOVOS_BASE_URL"),
				os.Getenv("SOVOS_PROCESS_MESSAGE_ENDPOINT"),
			)
			method := "POST"
			payload := fmt.Sprintf(`{
				"Receiver": "%s",
				"Sender": "%s",
				"MessageId": "%s"
			}`, result.Receiver, result.Sender, messageID)
			payloadReader := strings.NewReader(payload)

			_, err := makeRequest(method, url, payloadReader)
			if err != nil {
				log.Fatalf("Error with PROCESSQUEUEDMESSAGE call: %s", err)

				return false
			}

			return true
		}
	}

	return false
}

func getAuthHeader() string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(os.Getenv("SOVOS_API_KEY")+":"+os.Getenv("SOVOS_API_SECRET")))
}

func makeRequest(method string, url string, payload io.Reader) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return []byte(""), err
	}
	req.Header.Add("Authorization", getAuthHeader())
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return []byte(""), err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	return body, err
}

func clearAllMessages() {
	if !yesNo(fmt.Sprintf("Are you sure?  This will clear ALL messages currently on the %s queue", os.Getenv("SOVOS_ENVIRONMENT"))) {
		os.Exit(1)
	}

	for _, message := range getAllMessages().Results.Messages {
		if !processMessageById(message.ID) {
			fmt.Printf("Failed to clear message %s from the queue\n", message.ID)
		} else {
			fmt.Printf("Message %s cleared from the queue\n", message.ID)
		}
	}
}

func yesNo(label string) bool {
	prompt := promptui.Select{
		Label: label,
		Items: []string{"Yes", "No"},
	}
	_, result, err := prompt.Run()
	if err != nil {
		log.Fatalf("Prompt failed %v\n", err)
	}
	return result == "Yes"
}
