package better_cloudflare

import (
	"context"
	"math/rand"
	"net"
	"strings"

	"github.com/F-TD5X/coredns_plugins/geodata"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

var log = clog.NewWithPlugin(pluginName)

type BetterCloudflare struct {
	Next       plugin.Handler
	ipv4       []net.IP
	ipv6       []net.IP
	categories map[string]struct{}
}

type ResponseReverter struct {
	dns.ResponseWriter
	p *BetterCloudflare
}

func (w *ResponseReverter) WriteMsg(res1 *dns.Msg) error {
	res := res1.Copy()
	for _, item := range res.Answer {
		if t, ok := item.(*dns.A); ok {
			if w.p.Match(&t.A) {
				w.p.randomPickA(t)
			}
		}
		if t, ok := item.(*dns.AAAA); ok {
			if w.p.Match(&t.AAAA) {
				w.p.randomPickAAAA(t)
			}
		}
	}
	return w.ResponseWriter.WriteMsg(res)
}

func (w *ResponseReverter) Write(buf []byte) (int, error) {
	n, err := w.ResponseWriter.Write(buf)
	return n, err
}

func (p *BetterCloudflare) randomPickA(t *dns.A) {
	t.A = randomPick(p.ipv4)
}

func (p *BetterCloudflare) randomPickAAAA(t *dns.AAAA) {
	t.AAAA = randomPick(p.ipv6)
}

func randomPick(l []net.IP) net.IP {
	randomIndex := rand.Intn(len(l))
	return l[randomIndex]
}

func (p *BetterCloudflare) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}
	wr := &ResponseReverter{w, p}
	rcode, err := plugin.NextOrFailure(pluginName, p.Next, ctx, wr, r)
	if plugin.ClientWrite(rcode) {
		return rcode, err
	}
	res := new(dns.Msg).SetRcode(r, rcode)
	state.SizeAndDo(res)
	wr.WriteMsg(res)
	return dns.RcodeSuccess, err
}

func (p *BetterCloudflare) Metadata(ctx context.Context, state request.Request) context.Context {
	return ctx
}

func (p *BetterCloudflare) Match(ip *net.IP) bool {
	codes := geodata.GeoIPDB.LookupCode(*ip)
	for _, code := range codes {
		if _, ok := p.categories[code]; ok {
			return true
		}
	}
	return false
}

func (p *BetterCloudflare) Load(geoipCategories []string, ipv4 []net.IP, ipv6 []net.IP) {
	p.ipv4 = ipv4
	p.ipv6 = ipv6
	categoryLength := len("geoip:")
	for _, category := range geoipCategories {
		if j := strings.Index(category, "geoip:"); j > 0 {
			p.categories[category[j+categoryLength:]] = struct{}{}
		}
	}
}

func (p *BetterCloudflare) Name() string { return pluginName }
