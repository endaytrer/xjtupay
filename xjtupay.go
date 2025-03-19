package xjtupay

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

type MoneyAmount string

type PaymentInfo struct {
	TransactionAmount MoneyAmount `json:"tranamt"`
	Account           string      `json:"account"`
	Sno               string      `json:"sno"`
	ToAccount         string      `json:"toaccount"`
	ThirdSystem       string      `json:"thirdsystem"`
	ThirdOrderId      string      `json:"thirdorderid"`
	OrderType         string      `json:"ordertype"`
	Signature         string      `json:"sign"`
	OrderDesc         string      `json:"orderdesc"`
	Param1            string      `json:"praram1"`
	ThirdURL          string      `json:"thirdurl"`
}

func (t *PaymentInfo) toForm() url.Values {
	ans := make(url.Values)
	ans.Add("tranamt", string(t.TransactionAmount))
	ans.Add("account", t.Account)
	ans.Add("sno", t.Sno)
	ans.Add("toaccount", t.ToAccount)
	ans.Add("thirdsystem", t.ThirdSystem)
	ans.Add("thirdorderid", t.ThirdOrderId)
	ans.Add("ordertype", t.OrderType)
	ans.Add("sign", t.Signature)
	ans.Add("orderdesc", t.OrderDesc)
	ans.Add("praram1", t.Param1)
	ans.Add("thirdurl", t.ThirdURL)

	return ans
}

type PaymentError int

const (
	RequirePassword PaymentError = iota
	RequestError
	InitFailed
	PaymentFailed
	Unimplemented
)

func (t PaymentError) Error() string {
	switch t {
	case RequirePassword:
		return "Payment Error: Require Password"
	case RequestError:
		return "Payment Error: Request Status Error"
	case InitFailed:
		return "Payment Error: Initialization Failed"
	case PaymentFailed:
		return "Payment Error: Payment Failed"
	case Unimplemented:
		return "Payment Error: Unimplemented"
	}
	panic("not implemented")
}

var selector_payItem = cascadia.MustCompile(".payitem>li>span")
var selector_orderId = cascadia.MustCompile("#txtorderid")
var selector_passwordInput = cascadia.MustCompile("#password")
var selector_isNotice = cascadia.MustCompile("#form1>input[name=isnotice]")

var selector_postForm = cascadia.MustCompile("#myform")

type Passcode string

type generalPayment struct {
	client  http.Client
	headers http.Header

	orderId  string
	payId    string
	param1   string
	passwd   string
	payType  string
	isNotice string
}

func (t *generalPayment) toForm() url.Values {
	ans := make(url.Values)

	ans.Add("orderid", t.orderId)
	ans.Add("payid", t.payId)
	ans.Add("param1", t.param1)
	ans.Add("passwd", t.passwd)
	ans.Add("paytype", t.payType)
	ans.Add("isnotice", t.isNotice)

	return ans
}

func (t *generalPayment) pay(passcode Passcode) (*http.Client, error) {
	root, err := t.request("/Pay/CommonMobilePay", t.toForm())
	if err != nil {
		return nil, err
	}
	form := cascadia.Query(root, selector_postForm)
	if form == nil {
		return nil, PaymentFailed
	}

	post_url := getAttribute(form, "action")

	form_data := make(url.Values)

	for child := range form.ChildNodes() {
		key := getAttribute(child, "name")
		value := getAttribute(child, "value")
		form_data.Add(key, value)
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		panic("Cookie jar creation failed")
	}
	client := http.Client{Jar: jar}

	req, err := http.NewRequest(http.MethodPost, post_url, strings.NewReader(form_data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header = make(http.Header)
	for k, v := range t.headers {
		req.Header[k] = v
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	_, err = client.Do(req)
	return &client, nil
}

type PayWithPasswd generalPayment

func (t *PayWithPasswd) Pay(passcode Passcode) (*http.Client, error) {
	return t.pay(passcode)
}
func (t *PayWithPasswd) pay(passcode Passcode) (*http.Client, error) {
	gp := generalPayment(*t)
	return gp.pay(passcode)
}

type PayWithoutPasswd generalPayment

func (t *PayWithoutPasswd) Pay() (*http.Client, error) {
	return t.pay("")
}
func (t *PayWithoutPasswd) pay(passcode Passcode) (*http.Client, error) {
	gp := generalPayment(*t)
	return gp.pay(passcode)
}

type PayMethod interface {
	pay(passcode Passcode) (*http.Client, error)
}

const paymentUrl = "http://202.117.1.244:9001"

func (t *generalPayment) request(url string, values url.Values) (*html.Node, error) {
	value_encoded := values.Encode()
	req, err := http.NewRequest(http.MethodPost, paymentUrl+url, strings.NewReader(value_encoded))
	if err != nil {
		return nil, err
	}
	req.Header = make(http.Header)
	for k, v := range t.headers {
		req.Header[k] = v
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, RequestError
	}
	node, err := html.Parse(res.Body)
	if err != nil {
		return nil, err
	}
	return node, nil
}

func getAttribute(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}
func InitPayment(mobile bool, payment PaymentInfo) (PayMethod, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic("Cookie jar creation failed")
	}
	ans := generalPayment{
		client:  http.Client{Jar: jar},
		headers: http.Header{},
	}

	if mobile {
		ans.headers.Set("User-Agent", "Mozilla/5.0 (Linux; Android 14; 2211133C Build/UKQ1.230804.001; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/126.0.6478.134 Mobile Safari/537.36 toon/2123344193 toonType/150 toonVersion/6.4.0 toongine/1.0.12 toongineBuild/12 platform/android language/zh skin/white fontIndex/0")
		ans.headers.Set("X-Requested-With", "synjones.commerce.xjtu")
	} else {
		ans.headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
	}

	root, err := ans.request("/Order/CreateOrder", payment.toForm())
	if err != nil {
		return nil, err
	}

	node_payItem := cascadia.Query(root, selector_payItem)
	node_orderId := cascadia.Query(root, selector_orderId)
	node_passwordInput := cascadia.Query(root, selector_passwordInput)
	node_isNotice := cascadia.Query(root, selector_isNotice)

	if node_payItem == nil || node_orderId == nil || node_passwordInput == nil || node_isNotice == nil {
		return nil, InitFailed
	}

	ans.orderId = getAttribute(node_orderId, "value")
	ans.payId = getAttribute(node_payItem, "payid")
	ans.param1 = getAttribute(node_payItem, "tag")
	ans.payType = getAttribute(node_payItem, "intername")
	ans.isNotice = getAttribute(node_isNotice, "value")

	passwd_needed_style := getAttribute(node_passwordInput, "style")
	passwd_needed := !strings.Contains(passwd_needed_style, "display: none")
	var ans_deviated PayMethod
	if passwd_needed {
		t := PayWithPasswd(ans)
		ans_deviated = &t
	} else {
		t := PayWithoutPasswd(ans)
		ans_deviated = &t
	}
	return ans_deviated, nil
}
