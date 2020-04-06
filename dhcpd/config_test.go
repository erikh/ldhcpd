package dhcpd

import (
	"reflect"
	"testing"
	"time"
)

func TestConfig(t *testing.T) {
	outConfigs := map[string]Config{
		"basic": {
			LeaseDuration: defaultLeaseDuration,
			DNSServers: []string{
				"10.0.0.1",
				"1.1.1.1",
			},
			Gateway: "10.0.20.1",
			DynamicRange: Range{
				From: "10.0.20.50",
				To:   "10.0.20.100",
			},
			DBFile: defaultDBFile,
			Certificate: Certificate{
				CAFile:   defaultCAFile,
				CertFile: defaultCertFile,
				KeyFile:  defaultKeyFile,
			},
		},
		"no dns": {
			LeaseDuration: defaultLeaseDuration,
			DNSServers:    []string{},
			Gateway:       "10.0.20.1",
			DynamicRange: Range{
				From: "10.0.20.50",
				To:   "10.0.20.100",
			},
			DBFile: defaultDBFile,
			Certificate: Certificate{
				CAFile:   defaultCAFile,
				CertFile: defaultCertFile,
				KeyFile:  defaultKeyFile,
			},
		},
		"lease duration populated": {
			LeaseDuration: time.Hour,
			DNSServers: []string{
				"10.0.0.1",
				"1.1.1.1",
			},
			Gateway: "10.0.20.1",
			DynamicRange: Range{
				From: "10.0.20.50",
				To:   "10.0.20.100",
			},
			DBFile: defaultDBFile,
			Certificate: Certificate{
				CAFile:   defaultCAFile,
				CertFile: defaultCertFile,
				KeyFile:  defaultKeyFile,
			},
		},
		"db file populated": {
			LeaseDuration: defaultLeaseDuration,
			DNSServers: []string{
				"10.0.0.1",
				"1.1.1.1",
			},
			Gateway: "10.0.20.1",
			DynamicRange: Range{
				From: "10.0.20.50",
				To:   "10.0.20.100",
			},
			DBFile: "foo.db",
			Certificate: Certificate{
				CAFile:   defaultCAFile,
				CertFile: defaultCertFile,
				KeyFile:  defaultKeyFile,
			},
		},
		"cert configuration populated": {
			LeaseDuration: defaultLeaseDuration,
			DNSServers: []string{
				"10.0.0.1",
				"1.1.1.1",
			},
			Gateway: "10.0.20.1",
			DynamicRange: Range{
				From: "10.0.20.50",
				To:   "10.0.20.100",
			},
			DBFile: defaultDBFile,
			Certificate: Certificate{
				CAFile:   "cacert.pem",
				CertFile: "server.pem",
				KeyFile:  "server.key",
			},
		},
	}

	validConfigs := map[string]Config{
		"basic": {
			DNSServers: []string{
				"10.0.0.1",
				"1.1.1.1",
			},
			Gateway: "10.0.20.1",
			DynamicRange: Range{
				From: "10.0.20.50",
				To:   "10.0.20.100",
			},
		},
		"no dns": {
			Gateway: "10.0.20.1",
			DynamicRange: Range{
				From: "10.0.20.50",
				To:   "10.0.20.100",
			},
		},
		"db file populated": {
			DNSServers: []string{
				"10.0.0.1",
				"1.1.1.1",
			},
			Gateway: "10.0.20.1",
			DynamicRange: Range{
				From: "10.0.20.50",
				To:   "10.0.20.100",
			},
			DBFile: "foo.db",
		},
		"lease duration populated": {
			LeaseDuration: time.Hour,
			DNSServers: []string{
				"10.0.0.1",
				"1.1.1.1",
			},
			Gateway: "10.0.20.1",
			DynamicRange: Range{
				From: "10.0.20.50",
				To:   "10.0.20.100",
			},
		},
		"cert configuration populated": {
			DNSServers: []string{
				"10.0.0.1",
				"1.1.1.1",
			},
			Gateway: "10.0.20.1",
			DynamicRange: Range{
				From: "10.0.20.50",
				To:   "10.0.20.100",
			},
			Certificate: Certificate{
				CAFile:   "cacert.pem",
				CertFile: "server.pem",
				KeyFile:  "server.key",
			},
		},
	}

	invalidConfigs := map[string]Config{
		"empty fields": {},
		"bad dns": {
			DNSServers: []string{
				"acdef",
			},
			Gateway: "10.0.20.1",
			DynamicRange: Range{
				From: "10.0.20.50",
				To:   "10.0.20.100",
			},
		},
		"bad gateway": {
			DNSServers: []string{
				"10.0.0.1",
				"1.1.1.1",
			},
			Gateway: "abcdef",
			DynamicRange: Range{
				From: "10.0.20.50",
				To:   "10.0.20.100",
			},
		},
		"bad ips in range/from": {
			DNSServers: []string{
				"10.0.0.1",
				"1.1.1.1",
			},
			Gateway: "10.0.20.1",
			DynamicRange: Range{
				From: "abcdef",
				To:   "10.0.20.100",
			},
		},
		"bad ips in range/to": {
			DNSServers: []string{
				"10.0.0.1",
				"1.1.1.1",
			},
			Gateway: "10.0.20.1",
			DynamicRange: Range{
				From: "10.0.20.50",
				To:   "abcdef",
			},
		},
		"invalid parseable range": {
			DNSServers: []string{
				"10.0.0.1",
				"1.1.1.1",
			},
			Gateway: "10.0.20.1",
			DynamicRange: Range{
				From: "10.0.20.100",
				To:   "10.0.20.50",
			},
		},
	}

	for name, config := range validConfigs {
		if err := config.validateAndFix(); err != nil {
			t.Fatalf("[%v] Errored when expected success: %v", name, err)
		}

		if !reflect.DeepEqual(outConfigs[name], config) {
			t.Fatalf("[%v] Configurations were not equal", name)
		}
	}

	for name, config := range invalidConfigs {
		if err := config.validateAndFix(); err == nil {
			t.Fatalf("[%v] Invalid configuration did not error", name)
		}
	}
}
