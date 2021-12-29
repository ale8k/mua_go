package mua

import (
	"testing"
)

func TestNewMailClient(t *testing.T) {
	t.Run("it creates a mail client successfully", func(t *testing.T) {
		if got := NewMailClient("bob@bob.com", "hi"); got == nil {
			t.Errorf("NewMailClient() = %v, should not be nil!", got)
		}
	})
}

func TestMailClient_connectInsecure(t *testing.T) {
	t.Run("it initiates a generic tcp connection to the server successfully", func(t *testing.T) {
		client := NewMailClient("example@example.com", "") // It connects by default, but we check regardless
		if err := client.connectSMTPInsecure(); err != nil {
			t.Errorf("connectInsecure() failed to connect!")
		}
	})

}

func TestMailClient_UpgradeConnectionTLS(t *testing.T) {
	t.Run("it upgrades the connection successfully without any panic", func(t *testing.T) {
		client := NewMailClient("bob@bob.com", "hi")
		client.connectSMTPInsecure()
		client.upgradeSMTPConnectionTLS()
	})

}
