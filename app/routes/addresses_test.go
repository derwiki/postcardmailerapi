package routes

import "testing"

func TestAddressesListHandler(t *testing.T) {
	resp := OrdinaryResponse{}

	err := utils.TestHandlerUnMarshalResp("POST", "/login", "form", user, &resp)
	if err != nil {
		t.Errorf("TestLoginHandler: %v\n", err)
		return
	}

	if resp.Errno != "0" {
		t.Errorf("TestLoginHandler: response is not expected\n")
		return
	}
}
