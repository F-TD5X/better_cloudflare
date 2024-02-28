package better_cloudflare

import (
	"net"

	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

const pluginName = "better_cloudflare"

func init() {
	plugin.Register(pluginName, setup)
}

func setup(c *caddy.Controller) error {
	betterCloudflare, err := parse(c)

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		betterCloudflare.Next = next
		return betterCloudflare
	})
	return err
}

func parse(c *caddy.Controller) (*BetterCloudflare, error) {
	var geoipCategories []string
	var ipv4 []net.IP
	var ipv6 []net.IP
	for c.Next() {
		if !c.NextArg() {
			return nil, c.ArgErr()
		}
		geoipCategories = c.RemainingArgs()
		for c.NextBlock() {
			switch dir := c.Val(); dir {
			case "ipv4":
				args := c.RemainingArgs()
				for _, ip := range args {
					ipv4 = append(ipv4, net.ParseIP(ip))
				}
			case "ipv6":
				args := c.RemainingArgs()
				for _, ip := range args {
					ipv6 = append(ipv4, net.ParseIP(ip))
				}
			default:
				log.Error("invalid parameter in block.")
			}
		}
	}
	ret := &BetterCloudflare{}
	ret.Load(geoipCategories, ipv4, ipv6)
	return ret, nil
}
