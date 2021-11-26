package httplog

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	"golang.org/x/net/http2"
)

// H2c test
var H2c = http.Client{
	Transport: &http2.Transport{
		AllowHTTP: true,
		DialTLS:   Conn(context.Background()),
		TLSClientConfig: &tls.Config{
			// 跳过不可信证书检测
			InsecureSkipVerify: true, //  x509: certificate signed by unknown authority 由未知权威机构签署的证书
		},
	},
}

func TestH2cClient() {
	tr := &http2.Transport{
		AllowHTTP: true,
		DialTLS:   Conn(context.Background()),
		TLSClientConfig: &tls.Config{
			// 跳过不可信证书检测
			InsecureSkipVerify: true, //  x509: certificate signed by unknown authority 由未知权威机构签署的证书
		},
	}
	defer tr.CloseIdleConnections()
	c := http.Client{
		Transport: tr,
	}

	r, err := c.Get("http://localhost/events")
	if err != nil {
		// http: server gave HTTP response to HTTPS client
		panic(err)
	}

	defer r.Body.Close()

	b, _ := ioutil.ReadAll(r.Body)

	fmt.Println(string(b))
}

// h2,h2c
func Conn(ctx context.Context) func(network, addr string, cfg *tls.Config) (net.Conn, error) {

	return func(network, addr string, cfg *tls.Config) (net.Conn, error) {

		dialer := &tls.Dialer{Config: cfg}
		cn, err := dialer.DialContext(ctx, network, addr)
		if err != nil {
			//h2c http2: dialer context err: tls: first record does not look like a TLS handshake
			if err.Error() == "tls: first record does not look like a TLS handshake" {
				return net.Dial(network, addr)
			}
			log.Printf("http2: dialer context err: %s \n", err.Error())
			return nil, err
		}
		tlsCn := cn.(*tls.Conn)

		state := tlsCn.ConnectionState()
		if p := state.NegotiatedProtocol; p != http2.NextProtoTLS {
			// return net.Dial(network, addr)
			return nil, fmt.Errorf("http2: unexpected ALPN protocol %q; want %q", p, http2.NextProtoTLS)
		}
		if !state.NegotiatedProtocolIsMutual {
			return nil, errors.New("http2: could not negotiate protocol mutually")
		}
		return tlsCn, nil
	}

}
