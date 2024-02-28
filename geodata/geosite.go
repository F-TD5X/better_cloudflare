package geodata

import (
	"os"

	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/metacubex/geo/geoip"
	"github.com/metacubex/geo/geosite"
)

var (
	GeoSiteDB *geosite.Database
	GeoIPDB   *geoip.Database
)

func init() {
	var err error
	GeoSiteDB, err = geosite.FromFile("GeoSite.dat")
	if err != nil {
		clog.Fatal("unable to load GeoSite database.", err)
	}
	GeoIPDB, err = geoip.FromFile("GeoIP.dat")
	if err != nil {
		clog.Fatal("unable to load GeoIP database.", err)
	}
	if GeoSiteDB == nil || GeoIPDB == nil {
		os.Exit(-1)
	}
}
