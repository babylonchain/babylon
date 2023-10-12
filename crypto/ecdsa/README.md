# ECDSA

This module implements Bitcoin's ECDSA signature algorithm over Secp256k1 curve.
It follows the "message sign" format in [Bitcoin](https://github.com/bitcoin/bitcoin/pull/524).
The format is used in [OKX wallet SDK](https://github.com/okx/js-wallet-sdk/blob/a57c2acbe6ce917c0aa4e951d96c4e562ad58444/packages/coin-bitcoin/src/BtcWallet.ts#L331).

References:

- [Original design and implementation](https://github.com/bitcoin/bitcoin/pull/524)
- [An unofficial spec](https://github.com/fivepiece/sign-verify-message/blob/master/signverifymessage.md)
- [Implementation of OKX wallet SDK](https://github.com/okx/js-wallet-sdk/blob/a57c2acbe6ce917c0aa4e951d96c4e562ad58444/packages/coin-bitcoin/src/BtcWallet.ts#L331)
