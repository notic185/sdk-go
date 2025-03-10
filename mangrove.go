package sdk

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/elliotchance/pie/v2"
	"github.com/google/uuid"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Mangrove struct {
	credential UserCredential
	endpoint   *url.URL

	MerchantOrder *MangroveMerchantOrder
	Order         *MangroveOrder
	User          *MangroveUser
	UserOrder     *MangroveUserOrder
}

// MangroveMerchantOrder 请不要直接使用该类型，因为它随时都有可能会被重命名
type MangroveMerchantOrder struct {
	m *Mangrove
}

// MangroveOrder 请不要直接使用该类型，因为它随时都有可能会被重命名
type MangroveOrder struct {
	m *Mangrove
}

// MangroveUser 请不要直接使用该类型，因为它随时都有可能会被重命名
type MangroveUser struct {
	m *Mangrove
}

// MangroveUserOrder 请不要直接使用该类型，因为它随时都有可能会被重命名
type MangroveUserOrder struct {
	m *Mangrove
}

func New(endpoint string, credential UserCredential) (*Mangrove, error) {
	// -
	// > -
	parsedEndpoint, wrong := url.Parse(endpoint)
	// > -
	if wrong != nil {
		return nil, wrong
	}

	// -
	return (&Mangrove{
		endpoint:   parsedEndpoint,
		credential: credential,
	}).initialize(), nil
}

func (m *Mangrove) initialize() *Mangrove {
	// 初始化子客户端
	// > 商户订单
	m.MerchantOrder = &MangroveMerchantOrder{m}
	// > 订单
	m.Order = &MangroveOrder{m}
	// > 用户
	m.User = &MangroveUser{m}
	// > 用户订单
	m.UserOrder = &MangroveUserOrder{m}

	// 返回自身
	return m
}

func (m *Mangrove) request(requestMethod string, requestPath string, requestPayload any, responsePayload any) error {
	// 初始化上下文
	// > 请求地址
	// > > 解析请求地址
	requestURL, wrong := url.Parse(
		fmt.Sprintf("%s%s", m.endpoint, requestPath),
	)
	// > > 错误处理
	if wrong != nil {
		return wrong
	}
	// > 请求
	request := &http.Request{
		Method: requestMethod,
		URL:    requestURL,
		Header: http.Header{
			"content-type":        []string{"application/json"},
			"host":                []string{m.endpoint.Host},
			"x-guarder-id":        []string{m.credential.ExternalId},
			"x-guarder-signed-at": []string{strconv.FormatInt(time.Now().UnixMilli(), 10)},
			"x-guarder-uuid":      []string{uuid.New().String()},
		},
	}
	// > 请求体
	if requestPayload != nil {
		// -
		requestBody, wrong := json.Marshal(requestPayload)
		// -
		if wrong == nil {
			request.Body = io.NopCloser(bytes.NewReader(requestBody))
		} else {
			return wrong
		}
	}
	// > 请求头
	// > > -
	requestHeaderAuthorization, wrong := m.signRequest(request)
	// > > -
	if wrong == nil {
		request.Header["Authorization"] = []string{
			fmt.Sprintf("Signature %s", requestHeaderAuthorization),
		}
	} else {
		return wrong
	}

	// 执行请求
	// > > -
	response, wrong := http.DefaultClient.Do(request)
	// > > 错误处理
	if wrong != nil {
		return wrong
	}

	// 解析结果并返回
	// > 读取
	// > > -
	responseBody, wrong := io.ReadAll(response.Body)
	// > > -
	if wrong != nil {
		return wrong
	}
	// > 解析
	// > > -
	responsePayloadWrapper := &Result{}
	// > > -
	if wrong = json.Unmarshal(responseBody, responsePayloadWrapper); wrong != nil {
		return wrong
	}
	// > 处理
	if responsePayloadWrapper.Code == 200 {
		// -
		// > -
		rawResponsePayload, wrong := json.Marshal(responsePayloadWrapper.Data)
		// > -
		if wrong != nil {
			return wrong
		}

		// -
		return json.Unmarshal(rawResponsePayload, responsePayload)
	} else {
		return errors.New(responsePayloadWrapper.Message)
	}
}

func (m *Mangrove) signRequest(request *http.Request) (string, error) {
	// 构建签名载荷
	// > -
	var signaturePayload []string
	// > -
	// > > 请求方法和请求路径
	signaturePayload = append(
		signaturePayload, fmt.Sprintf("%s %s", request.Method, request.URL.Path),
	)
	// > > 请求头
	for _, key := range []string{"content-type", "host", "x-guarder-id", "x-guarder-signed-at", "x-guarder-uuid"} {
		for key0, value := range request.Header {
			if strings.ToLower(key0) == key {
				signaturePayload = append(
					signaturePayload, fmt.Sprintf("%s: %s", key, value[0]),
				)
			}
		}
	}
	// > > 空行
	signaturePayload = append(signaturePayload, "")
	// > > 请求体
	if request.Body == nil {
		signaturePayload = append(signaturePayload, "")
	} else {
		// 读取请求体
		// > -
		requestBody, wrong := io.ReadAll(request.Body)
		// > -
		if wrong == nil {
			request.Body = io.NopCloser(bytes.NewReader(requestBody))
		} else {
			return "", wrong
		}

		// 解析请求体
		// > -
		var parsedRequestBody any
		// > -
		if wrong := json.Unmarshal(requestBody, &parsedRequestBody); wrong != nil {
			return "", wrong
		}

		// 处理请求体
		// > -
		transformedRequestBody := u.PileDown(parsedRequestBody, "")
		// > -
		signaturePayload = append(
			signaturePayload,
			pie.Join(
				pie.Map(
					pie.Sort(
						pie.Keys(transformedRequestBody),
					),
					func(key string) string {
						if transformedRequestBody[key] == nil {
							return fmt.Sprintf("%s=null", key)
						} else {
							return fmt.Sprintf("%s=%v", key, transformedRequestBody[key])
						}
					},
				),
				"&",
			),
		)
	}

	// 签名并返回
	// > -
	signature := hmac.New(sha512.New, []byte(m.credential.Secret))
	// > -
	signature.Write([]byte(strings.Join(signaturePayload, "\r\n")))
	// > -
	return hex.EncodeToString(signature.Sum(nil)), nil
}

func (m *MangroveMerchantOrder) Create(merchantOrders []MerchantOrder) (merchantOrders0 []MerchantOrder, wrong error) {
	// -
	wrong = m.m.request(
		"PUT", "/v1.2/merchant-order", merchantOrders, &merchantOrders0,
	)
	// -
	return
}

func (m *MangroveOrder) Describe(uuid0 string) (order Order, wrong error) {
	// -
	wrong = m.m.request(
		"GET", fmt.Sprintf("/v1.2/order/%s", uuid0), nil, &order,
	)
	// -
	return
}

func (m *MangroveOrder) Update(order Order) (orders []Order, wrong error) {
	// -
	wrong = m.m.request(
		"PATCH", fmt.Sprintf("/v1.2/order/%s", order.UUID), order, &orders,
	)

	// -
	return
}

func (m *MangroveOrder) Delete(uuid0 string) (wrong error) {
	// -
	wrong = m.m.request(
		"DELETE", fmt.Sprintf("/v1.2/order/%s", uuid0), nil, nil,
	)
	// -
	return
}

func (m *MangroveOrder) HandleCallback(request *http.Request) (order Order, wrong error) {
	// 补全信息
	request.Header["Host"] = []string{request.Host}

	// 构建签名
	// > -
	requestSignature, wrong := m.m.signRequest(request)
	// > -
	if wrong != nil {
		return
	}

	// 验证签名
	// > -
	requestHeaderAuthorization := request.Header["Authorization"][0]
	// > -
	if requestSignature == strings.Split(requestHeaderAuthorization, " ")[1] {
		// -
		// > -
		requestBody, wrong0 := io.ReadAll(request.Body)
		// > -
		if wrong0 != nil {
			// -
			wrong = wrong0
			// -
			return
		}

		// -
		wrong = json.Unmarshal(requestBody, &order)
	} else {
		wrong = errors.New("invalid signature")
	}

	// 返回
	return
}

func (m *MangroveUserOrder) Create(userOrders []UserOrder) (userOrders0 []UserOrder, wrong error) {
	// -
	wrong = m.m.request(
		"PUT", "/v1.2/user-order", userOrders, &userOrders0,
	)
	// -
	return
}

func (m *MangroveUser) SummarizeIntegralAmount() (result map[string]float64, wrong error) {
	// -
	wrong = m.m.request(
		"GET", "/v1.2/user/summarize-integral-amount", nil, &result,
	)
	// -
	return
}
