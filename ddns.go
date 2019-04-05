package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"golang.org/x/net/publicsuffix"
)

func must(err error) {
	if err == nil {
		return
	}

	log.Fatal(err)
}

func main() {
	silent := flag.Bool("s", false, "suppress output on success")
	ttl := flag.Int("ttl", 120, "the TTL value to use if creating any new records")
	flag.Parse()

	domainName := flag.Arg(0)
	if domainName == "" {
		log.Fatal("missing domain name argument")
	}

	zoneName, err := publicsuffix.EffectiveTLDPlusOne(domainName)
	must(err)

	ip4, err := httpGetIP("https://ip4.seeip.org")
	must(err)

	ip6, err := httpGetIP("https://ip6.seeip.org")
	must(err)

	api, err := cloudflare.New(os.Getenv("CF_API_KEY"), os.Getenv("CF_API_EMAIL"))
	must(err)

	zoneID, err := api.ZoneIDByName(zoneName)
	must(err)

	rrs, err := api.DNSRecords(zoneID, cloudflare.DNSRecord{
		Type: "A,AAAA",
		Name: domainName,
	})
	must(err)

	for _, r := range rrs {
		switch r.Type {
		case "A":
			r.Content, ip4 = ip4, ""
		case "AAAA":
			r.Content, ip6 = ip6, ""
		default:
			log.Fatalf("unexpected record type returned (id=%s): %s", r.ID, r.Type)
		}

		if r.Content == "" {
			must(api.DeleteDNSRecord(zoneID, r.ID))
		} else {
			must(api.UpdateDNSRecord(zoneID, r.ID, r))
		}
	}

	for typ, ip := range map[string]string{
		"A":    ip4,
		"AAAA": ip6,
	} {
		if ip == "" {
			continue
		}

		_, err := api.CreateDNSRecord(zoneID, cloudflare.DNSRecord{
			Type:    typ,
			Name:    domainName,
			Content: ip,
			TTL:     *ttl,
		})
		must(err)
	}

	if *silent {
		return
	}

	rrs, err = api.DNSRecords(zoneID, cloudflare.DNSRecord{
		Type: "A,AAAA",
		Name: domainName,
	})
	must(err)

	for _, r := range rrs {
		fmt.Printf("%s %-4s %s\n", r.Name, r.Type, r.Content)
	}
}

func httpGetIP(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		if strings.Contains(err.Error(), "connect: network is unreachable") {
			// IPv6 requests on IPv4-only host will return this
			// error which we ignore.
			return "", nil
		}

		return "", err
	}
	defer resp.Body.Close()

	body := io.LimitReader(resp.Body, 1<<10)
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return "", err
	}

	ip := string(bytes.TrimSpace(b))
	if net.ParseIP(ip) == nil {
		return "", fmt.Errorf("%s returned invalid ip address: %q", url, ip)
	}

	return ip, nil
}
