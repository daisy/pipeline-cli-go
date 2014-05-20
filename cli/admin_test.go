package cli

import "testing"

//Test delete client command output
func TestDeleteClient(t *testing.T) {
	cli, link, _ := makeReturningCli(nil, t)
	r := overrideOutput(cli)
	cli.AddDeleteClientCommand(link)
	err := cli.Run([]string{"delete", "id"})
	if err != nil {
		t.Errorf("Unexpected error %v", err)
	}
	if getCall(link) != DELETE_CLIENT_CALL {
		t.Errorf("delete client wasn't called")
	}

	result := string(r.Bytes())
	expected := "Client id deleted\n"
	if result != string(expected) {
		t.Errorf("client delete error %s!=%s", string(expected), result)
	}

}

//Test delete client command id check
func TestDeleteClientNoId(t *testing.T) {
	cli, link, _ := makeReturningCli(nil, t)
	//r := overrideOutput(cli)
	cli.AddDeleteClientCommand(link)
	err := cli.Run([]string{"delete"})
	if err == nil {
		t.Errorf("Delete client needs an id")
	}
}

//Test delete client command id check
func TestDeleteClientError(t *testing.T) {
	cli, link, pipe := makeReturningCli(nil, t)
	pipe.failOnCall = DELETE_CLIENT_CALL
	//r := overrideOutput(cli)
	cli.AddDeleteClientCommand(link)
	err := cli.Run([]string{"delete", "nonexistent id"})
	if getCall(link) != DELETE_CLIENT_CALL {
		t.Errorf("delete client wasn't called")
	}
	if err == nil {
		t.Errorf("Link error")
	}
}
