package cidranger

import (
	"fmt"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	rnet "github.com/yl2chen/cidranger/net"
)

func TestPrefixTrieInsert(t *testing.T) {
	cases := []struct {
		version                      rnet.IPVersion
		inserts                      []string
		expectedNetworksInDepthOrder []string
		name                         string
	}{
		{rnet.IPv4, []string{"192.168.0.1/24"}, []string{"192.168.0.1/24"}, "basic insert"},
		{
			rnet.IPv4,
			[]string{"1.2.3.4/32", "1.2.3.5/32"},
			[]string{"1.2.3.4/32", "1.2.3.5/32"},
			"single ip IPv4 network insert",
		},
		{
			rnet.IPv6,
			[]string{"0::1/128", "0::2/128"},
			[]string{"0::1/128", "0::2/128"},
			"single ip IPv6 network insert",
		},
		{
			rnet.IPv4,
			[]string{"192.168.0.1/16", "192.168.0.1/24"},
			[]string{"192.168.0.1/16", "192.168.0.1/24"},
			"in order insert",
		},
		{
			rnet.IPv4,
			[]string{"192.168.0.1/32", "192.168.0.1/32"},
			[]string{"192.168.0.1/32"},
			"duplicate network insert",
		},
		{
			rnet.IPv4,
			[]string{"192.168.0.1/24", "192.168.0.1/16"},
			[]string{"192.168.0.1/16", "192.168.0.1/24"},
			"reverse insert",
		},
		{
			rnet.IPv4,
			[]string{"192.168.0.1/24", "192.168.1.1/24"},
			[]string{"192.168.0.1/24", "192.168.1.1/24"},
			"branch insert",
		},
		{
			rnet.IPv4,
			[]string{"192.168.0.1/24", "192.168.1.1/24", "192.168.1.1/30"},
			[]string{"192.168.0.1/24", "192.168.1.1/24", "192.168.1.1/30"},
			"branch inserts",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			trie := newPrefixTree(tc.version).(*prefixTrie)
			for _, insert := range tc.inserts {
				_, network, _ := net.ParseCIDR(insert)
				err := trie.Insert(NewBasicRangerEntry(*network))
				assert.NoError(t, err)
			}
			walk := trie.walkDepth()
			for _, network := range tc.expectedNetworksInDepthOrder {
				_, ipnet, _ := net.ParseCIDR(network)
				expected := NewBasicRangerEntry(*ipnet)
				actual := <-walk
				assert.Equal(t, expected, actual)
			}

			// Ensure no unexpected elements in trie.
			for network := range walk {
				assert.Nil(t, network)
			}
		})
	}
}

// @VENKAT: Increase/Decrease Route Scale
const (
	TEST_BENCHMARK_SCALE_NUMBER = 50000
)

var (
	v4prefix  []*net.IPNet
	v6prefix  []*net.IPNet
)

func createPatriciaV4Keys() {
	// insert v4 route entries
	intByt2 := 1
	intByt3 := 1
	intByt1 := 22
	byte4 := "0/24"
	for n := uint32(0); n < uint32(TEST_BENCHMARK_SCALE_NUMBER); n++ {
		if intByt2 > 253 && intByt3 > 254 {
			intByt1++
			intByt2 = 1
			intByt3 = 1

		}
		if intByt3 > 254 {
			intByt3 = 1
			intByt2++

		} else {
			intByt3++

		}
		if intByt2 > 254 {
			intByt2 = 1

		}
		byte1 := strconv.Itoa(intByt1)
		byte2 := strconv.Itoa(intByt2)
		byte3 := strconv.Itoa(intByt3)
		rtNet := byte1 + "." + byte2 + "." + byte3 + "." + byte4
		_, network, _ := net.ParseCIDR(rtNet)

		// @VENKAT:: Uncomment this to print ip addresses
		//fmt.Println("rtNet: ", rtNet, network)
		v4prefix = append(v4prefix, network)

	}
	fmt.Println("V4 Prefix set:", len(v4prefix), "SubnetMask: ", byte4)
}

func TestPrefixTrieV4String(t *testing.T) {
	createPatriciaV4Keys()
	//inserts := []string{"192.168.0.1/24", "192.168.1.1/24", "192.168.1.1/30"}
	var contains bool
	startTime := time.Now()
	trie := newPrefixTree(rnet.IPv4).(*prefixTrie)
	for _, pfx := range v4prefix {
		//_, network, _ := net.ParseCIDR(insert)
		contains, _ = trie.Contains(pfx.IP)
		if contains == false {
			trie.Insert(NewBasicRangerEntry(*pfx))

		}
	}
	timeNow := time.Now()
	elapsedTime := timeNow.Sub(startTime)
	fmt.Println("elapsedTime V4Prefix Inserts:", elapsedTime, TEST_BENCHMARK_SCALE_NUMBER)
	/*
	   	expected := `0.0.0.0/0 (target_pos:31:has_entry:false)
	   | 1--> 192.168.0.0/23 (target_pos:8:has_entry:false)
	   | | 0--> 192.168.0.0/24 (target_pos:7:has_entry:true)
	   | | 1--> 192.168.1.0/24 (target_pos:7:has_entry:true)
	   | | | 0--> 192.168.1.0/30 (target_pos:1:has_entry:true)`
	   	assert.Equal(t, expected, trie.String())
	*/

	// @VENKAT:: Uncomment this to print the trie
	//fmt.Println("trie: ", trie.String())
	
	startTime = time.Now()
	for _, pfx := range v4prefix {
	    contains, _ := trie.Contains(pfx.IP)
		if contains == true {
			_, err := trie.Remove(*pfx)
			if err != nil {
			    fmt.Println("Failed to remove: ", *pfx)
			}
		}
	}
	timeNow = time.Now()
	elapsedTime = timeNow.Sub(startTime)
	fmt.Println("elapsedTime V4Prefix Remove:", elapsedTime, TEST_BENCHMARK_SCALE_NUMBER)

	// @VENKAT:: Uncomment this to print the trie
	//fmt.Println("v4trie After Remove: ", trie.String())
}

func createPatriciaV6Keys() {
	// insert v6 route entries
	intByt2 := 10
	intByt3 := 10
	intByt1 := 10
	byte4 := "0/128"
	for n := uint32(0); n < uint32(TEST_BENCHMARK_SCALE_NUMBER); n++ {
		if intByt2 > 253 && intByt3 > 254 {
			intByt1++
			intByt2 = 1
			intByt3 = 1

		}
		if intByt3 > 254 {
			intByt3 = 1
			intByt2++

		} else {
			intByt3++

		}
		if intByt2 > 254 {
			intByt2 = 1

		}

		byte1 := strconv.Itoa(intByt1)
		byte2 := strconv.Itoa(intByt2)
		byte3 := strconv.Itoa(intByt3)
		rtNet := byte1 + ":" + byte2 + ":" + byte3 + "::" + byte4
		_, network, _ := net.ParseCIDR(rtNet)
		if network == nil {
		    fmt.Println("Invalid IP: ", rtNet)
		    panic(network)
		}

		// @venkat:: Uncomment this to print ip addresses
		//fmt.Println("rtNet: ", rtNet, network)
		v6prefix = append(v6prefix, network)

	}
	fmt.Println("V6 Prefix set:", len(v6prefix), "SubnetMask: ", byte4)
}

func TestPrefixTrieV6String(t *testing.T) {
	createPatriciaV6Keys()
	//inserts := []string{"192.168.0.1/24", "192.168.1.1/24", "192.168.1.1/30"}
	startTime := time.Now()
	trie := newPrefixTree(rnet.IPv6).(*prefixTrie)
	for _, pfx := range v6prefix {
	    contains, _ := trie.Contains(pfx.IP)
		if contains == false {
			trie.Insert(NewBasicRangerEntry(*pfx))

		}
	}
	timeNow := time.Now()
	elapsedTime := timeNow.Sub(startTime)
	fmt.Println("elapsedTime V6Prefix Inserts:", elapsedTime, TEST_BENCHMARK_SCALE_NUMBER)

	/*
	   	expected := `0.0.0.0/0 (target_pos:31:has_entry:false)
	   | 1--> 192.168.0.0/23 (target_pos:8:has_entry:false)
	   | | 0--> 192.168.0.0/24 (target_pos:7:has_entry:true)
	   | | 1--> 192.168.1.0/24 (target_pos:7:has_entry:true)
	   | | | 0--> 192.168.1.0/30 (target_pos:1:has_entry:true)`
	   	assert.Equal(t, expected, trie.String())
	*/
	// @venkat:: uncomment this to print the trie
	//fmt.Println("trie: ", trie.String())
	
	startTime = time.Now()
	for _, pfx := range v6prefix {
	    contains, _ := trie.Contains(pfx.IP)
		if contains == true {
			trie.Remove(*pfx)
		}
	}
	timeNow = time.Now()
	elapsedTime = timeNow.Sub(startTime)
	fmt.Println("elapsedTime V6Prefix Remove:", elapsedTime, TEST_BENCHMARK_SCALE_NUMBER)

	// @venkat:: uncomment this to print the trie
	//fmt.Println("v6trie After Remove: ", trie.String())
}

func TestPrefixTrieOperations(t *testing.T) {
    TestPrefixTrieV4String(t)
    TestPrefixTrieV6String(t)
}

func TestPrefixTrieRemove(t *testing.T) {
	cases := []struct {
		version                      rnet.IPVersion
		inserts                      []string
		removes                      []string
		expectedRemoves              []string
		expectedNetworksInDepthOrder []string
		name                         string
	}{
		{
			rnet.IPv4,
			[]string{"192.168.0.1/24"},
			[]string{"192.168.0.1/24"},
			[]string{"192.168.0.1/24"},
			[]string{},
			"basic remove",
		},
		{
			rnet.IPv4,
			[]string{"1.2.3.4/32", "1.2.3.5/32"},
			[]string{"1.2.3.5/32"},
			[]string{"1.2.3.5/32"},
			[]string{"1.2.3.4/32"},
			"single ip IPv4 network remove",
		},
		{
			rnet.IPv4,
			[]string{"0::1/128", "0::2/128"},
			[]string{"0::2/128"},
			[]string{"0::2/128"},
			[]string{"0::1/128"},
			"single ip IPv6 network remove",
		},
		{
			rnet.IPv4,
			[]string{"192.168.0.1/24", "192.168.0.1/25", "192.168.0.1/26"},
			[]string{"192.168.0.1/25"},
			[]string{"192.168.0.1/25"},
			[]string{"192.168.0.1/24", "192.168.0.1/26"},
			"remove path prefix",
		},
		{
			rnet.IPv4,
			[]string{"192.168.0.1/24", "192.168.0.1/25", "192.168.0.64/26", "192.168.0.1/26"},
			[]string{"192.168.0.1/25"},
			[]string{"192.168.0.1/25"},
			[]string{"192.168.0.1/24", "192.168.0.1/26", "192.168.0.64/26"},
			"remove path prefix with more than 1 children",
		},
		{
			rnet.IPv4,
			[]string{"192.168.0.1/24", "192.168.0.1/25"},
			[]string{"192.168.0.1/26"},
			[]string{""},
			[]string{"192.168.0.1/24", "192.168.0.1/25"},
			"remove non existent",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			trie := newPrefixTree(tc.version).(*prefixTrie)
			for _, insert := range tc.inserts {
				_, network, _ := net.ParseCIDR(insert)
				err := trie.Insert(NewBasicRangerEntry(*network))
				assert.NoError(t, err)
			}
			for i, remove := range tc.removes {
				_, network, _ := net.ParseCIDR(remove)
				removed, err := trie.Remove(*network)
				assert.NoError(t, err)
				if str := tc.expectedRemoves[i]; str != "" {
					_, ipnet, _ := net.ParseCIDR(str)
					expected := NewBasicRangerEntry(*ipnet)
					assert.Equal(t, expected, removed)
				} else {
					assert.Nil(t, removed)
				}
			}
			walk := trie.walkDepth()
			for _, network := range tc.expectedNetworksInDepthOrder {
				_, ipnet, _ := net.ParseCIDR(network)
				expected := NewBasicRangerEntry(*ipnet)
				actual := <-walk
				assert.Equal(t, expected, actual)
			}

			// Ensure no unexpected elements in trie.
			for network := range walk {
				assert.Nil(t, network)
			}
		})
	}
}

func TestToReplicateIssue(t *testing.T) {
	cases := []struct {
		version  rnet.IPVersion
		inserts  []string
		ip       net.IP
		networks []string
		name     string
	}{
		{
			rnet.IPv4,
			[]string{"192.168.0.1/32"},
			net.ParseIP("192.168.0.1"),
			[]string{"192.168.0.1/32"},
			"basic containing network for /32 mask",
		},
		{
			rnet.IPv6,
			[]string{"a::1/128"},
			net.ParseIP("a::1"),
			[]string{"a::1/128"},
			"basic containing network for /128 mask",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			trie := newPrefixTree(tc.version)
			for _, insert := range tc.inserts {
				_, network, _ := net.ParseCIDR(insert)
				err := trie.Insert(NewBasicRangerEntry(*network))
				assert.NoError(t, err)
			}
			expectedEntries := []RangerEntry{}
			for _, network := range tc.networks {
				_, net, _ := net.ParseCIDR(network)
				expectedEntries = append(expectedEntries, NewBasicRangerEntry(*net))
			}
			contains, err := trie.Contains(tc.ip)
			assert.NoError(t, err)
			assert.True(t, contains)
			networks, err := trie.ContainingNetworks(tc.ip)
			assert.NoError(t, err)
			assert.Equal(t, expectedEntries, networks)
		})
	}
}

type expectedIPRange struct {
	start net.IP
	end   net.IP
}

func TestPrefixTrieContains(t *testing.T) {
	cases := []struct {
		version     rnet.IPVersion
		inserts     []string
		expectedIPs []expectedIPRange
		name        string
	}{
		{
			rnet.IPv4,
			[]string{"192.168.0.0/24"},
			[]expectedIPRange{
				{net.ParseIP("192.168.0.0"), net.ParseIP("192.168.1.0")},
			},
			"basic contains",
		},
		{
			rnet.IPv4,
			[]string{"192.168.0.0/24", "128.168.0.0/24"},
			[]expectedIPRange{
				{net.ParseIP("192.168.0.0"), net.ParseIP("192.168.1.0")},
				{net.ParseIP("128.168.0.0"), net.ParseIP("128.168.1.0")},
			},
			"multiple ranges contains",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			trie := newPrefixTree(tc.version)
			for _, insert := range tc.inserts {
				_, network, _ := net.ParseCIDR(insert)
				err := trie.Insert(NewBasicRangerEntry(*network))
				assert.NoError(t, err)
			}
			for _, expectedIPRange := range tc.expectedIPs {
				var contains bool
				var err error
				start := expectedIPRange.start
				for ; !expectedIPRange.end.Equal(start); start = rnet.NextIP(start) {
					contains, err = trie.Contains(start)
					assert.NoError(t, err)
					assert.True(t, contains)
				}

				// Check out of bounds ips on both ends
				contains, err = trie.Contains(rnet.PreviousIP(expectedIPRange.start))
				assert.NoError(t, err)
				assert.False(t, contains)
				contains, err = trie.Contains(rnet.NextIP(expectedIPRange.end))
				assert.NoError(t, err)
				assert.False(t, contains)
			}
		})
	}
}

func TestPrefixTrieContainingNetworks(t *testing.T) {
	cases := []struct {
		version  rnet.IPVersion
		inserts  []string
		ip       net.IP
		networks []string
		name     string
	}{
		{
			rnet.IPv4,
			[]string{"192.168.0.0/24"},
			net.ParseIP("192.168.0.1"),
			[]string{"192.168.0.0/24"},
			"basic containing networks",
		},
		{
			rnet.IPv4,
			[]string{"192.168.0.0/24", "192.168.0.0/25"},
			net.ParseIP("192.168.0.1"),
			[]string{"192.168.0.0/24", "192.168.0.0/25"},
			"inclusive networks",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			trie := newPrefixTree(tc.version)
			for _, insert := range tc.inserts {
				_, network, _ := net.ParseCIDR(insert)
				err := trie.Insert(NewBasicRangerEntry(*network))
				assert.NoError(t, err)
			}
			expectedEntries := []RangerEntry{}
			for _, network := range tc.networks {
				_, net, _ := net.ParseCIDR(network)
				expectedEntries = append(expectedEntries, NewBasicRangerEntry(*net))
			}
			networks, err := trie.ContainingNetworks(tc.ip)
			assert.NoError(t, err)
			assert.Equal(t, expectedEntries, networks)
		})
	}
}

type coveredNetworkTest struct {
	version  rnet.IPVersion
	inserts  []string
	search   string
	networks []string
	name     string
}

var coveredNetworkTests = []coveredNetworkTest{
	{
		rnet.IPv4,
		[]string{"192.168.0.0/24"},
		"192.168.0.0/16",
		[]string{"192.168.0.0/24"},
		"basic covered networks",
	},
	{
		rnet.IPv4,
		[]string{"192.168.0.0/24"},
		"10.1.0.0/16",
		nil,
		"nothing",
	},
	{
		rnet.IPv4,
		[]string{"192.168.0.0/24", "192.168.0.0/25"},
		"192.168.0.0/16",
		[]string{"192.168.0.0/24", "192.168.0.0/25"},
		"multiple networks",
	},
	{
		rnet.IPv4,
		[]string{"192.168.0.0/24", "192.168.0.0/25", "192.168.0.1/32"},
		"192.168.0.0/16",
		[]string{"192.168.0.0/24", "192.168.0.0/25", "192.168.0.1/32"},
		"multiple networks 2",
	},
	{
		rnet.IPv4,
		[]string{"192.168.1.1/32"},
		"192.168.0.0/16",
		[]string{"192.168.1.1/32"},
		"leaf",
	},
	{
		rnet.IPv4,
		[]string{"0.0.0.0/0", "192.168.1.1/32"},
		"192.168.0.0/16",
		[]string{"192.168.1.1/32"},
		"leaf with root",
	},
	{
		rnet.IPv4,
		[]string{
			"0.0.0.0/0", "192.168.0.0/24", "192.168.1.1/32",
			"10.1.0.0/16", "10.1.1.0/24",
		},
		"192.168.0.0/16",
		[]string{"192.168.0.0/24", "192.168.1.1/32"},
		"path not taken",
	},
	{
		rnet.IPv4,
		[]string{
			"192.168.0.0/15",
		},
		"192.168.0.0/16",
		nil,
		"only masks different",
	},
}

func TestPrefixTrieCoveredNetworks(t *testing.T) {
	for _, tc := range coveredNetworkTests {
		t.Run(tc.name, func(t *testing.T) {
			trie := newPrefixTree(tc.version)
			for _, insert := range tc.inserts {
				_, network, _ := net.ParseCIDR(insert)
				err := trie.Insert(NewBasicRangerEntry(*network))
				assert.NoError(t, err)
			}
			var expectedEntries []RangerEntry
			for _, network := range tc.networks {
				_, net, _ := net.ParseCIDR(network)
				expectedEntries = append(expectedEntries,
					NewBasicRangerEntry(*net))
			}
			_, snet, _ := net.ParseCIDR(tc.search)
			networks, err := trie.CoveredNetworks(*snet)
			assert.NoError(t, err)
			assert.Equal(t, expectedEntries, networks)
		})
	}
}
