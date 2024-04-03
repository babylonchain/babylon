package client

import (
	"fmt"
	"path"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/juju/fslock"
)

func (c *Client) GetAddr() (string, error) {
	return c.provider.Address()
}

func (c *Client) MustGetAddr() string {
	addr, err := c.provider.Address()
	if err != nil {
		panic(fmt.Errorf("failed to get signer: %v", err))
	}
	return addr
}

func (c *Client) GetKeyring() keyring.Keyring {
	return c.provider.Keybase
}

// accessKeyWithLock triggers a function that access key ring while acquiring
// the file system lock, in order to remain thread-safe when multiple concurrent
// relayers are running on the same machine and accessing the same keyring
// adapted from
// https://github.com/babylonchain/babylon-relayer/blob/f962d0940832a8f84f747c5d9cbc67bc1b156386/bbnrelayer/utils.go#L212
func (c *Client) accessKeyWithLock(accessFunc func()) error {
	// use lock file to guard concurrent access to the keyring
	lockFilePath := path.Join(c.provider.PCfg.KeyDirectory, "keys.lock")
	lock := fslock.New(lockFilePath)
	if err := lock.Lock(); err != nil {
		return fmt.Errorf("failed to acquire file system lock (%s): %w", lockFilePath, err)
	}

	// trigger function that access keyring
	accessFunc()

	// unlock and release access
	if err := lock.Unlock(); err != nil {
		return fmt.Errorf("error unlocking file system lock (%s), please manually delete", lockFilePath)
	}

	return nil
}
