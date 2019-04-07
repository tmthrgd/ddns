# ddns

**ddns** is a dynamic DNS tool for [Cloudflare](https://www.cloudflare.com/) domains. It sets A and AAAA records for a domain to your systems public IP addresses as appropriate.

## Installation

Installation is simple and no different to any Go tool. The only requirement is a working [Go](https://golang.org/) install.

```
go get tmthrgd.dev/go/ddns
```

## Usage

Usage is simple with the `ddns` command taking the domain name to set.

```
ddns host.example.com
```

You need to set the `CF_API_EMAIL` and `CF_API_KEY` environment variables to your Cloudflare email address and API key which can be found at the bottom of the ["My Account" page](https://dash.cloudflare.com/profile).

## Notes

[SeeIP](https://seeip.org/) is used to resolve your systems public IPv4 and IPv6 addresses.