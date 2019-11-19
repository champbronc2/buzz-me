# buzz-me
BuzzMe is a service that allows you to create a profile and charge a small amount of BTC to receive messages.
Powered by the Lightning Network.

Inbound liquidity is provided by BitRefill Thor Lightning Channels

To run locally, you'll need to perform the following:
1. Install MongoDB
2. Install golang 1.13+
3. Download/install repository by running 'go get -u github.com/champbronc2/buzz-me'
4. Install bitcoind + LND
5. Configure the macaroon inside of lightning/lightning.go

A live version running on BTC mainnet can be found at http://ec2-18-237-219-223.us-west-2.compute.amazonaws.com:1323/ - Sign up, post, play!

Your profile will be visible at http://ec2-18-237-219-223.us-west-2.compute.amazonaws.com:1323/{username}

Completed functionality:
- Registartion
- Login
- Create a post, and be presented with a Lightning Network invoice
- View dashboard, and request withdrawals via Lightning Network. Performs validation that invoice is valid.

Pending functionality:
- Automatically detect if invoices are paid
- Automatically process withdrawals
- Automatically manage inbound liqiduity via Thor channels
- Allow profile to be updated
- Input sanitation
