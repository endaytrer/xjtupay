package xjtupay_test

import (
	"net/http"
	"testing"

	"github.com/endaytrer/xjtupay"
)

func TestMain(_ *testing.T) {
	payment, err := xjtupay.InitPayment(true, xjtupay.PaymentInfo{})

	if err != nil {
		panic(err.Error())
	}

	var redir_client *http.Client

	switch t := payment.(type) {

	case *xjtupay.PayWithPasswd:
		redir_client, err = t.Pay("888888")

	case *xjtupay.PayWithoutPasswd:
		redir_client, err = t.Pay()
	}
	if err != nil {
		panic(err.Error())
	}

	redir_client.CloseIdleConnections()
}
