// Copyright (c) 2022 Tigera, Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ut_test

import (
	"fmt"
	"net"
	"testing"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	. "github.com/onsi/gomega"
)

type packetTest struct {
	Description string
	IPv4Header  *layers.IPv4
	NextHeader  gopacket.Layer
	Payload     []byte
	Size        uint16
}

var malformedTestCases = []packetTest{
	{
		Description: "1 - A packet with IHL=4",
		IPv4Header: &layers.IPv4{
			Version: 4,
			IHL:     4,
			TTL:     64,
			Flags:   layers.IPv4DontFragment,
			SrcIP:   net.IPv4(4, 4, 4, 4),
			DstIP:   net.IPv4(1, 1, 1, 1),
		},
		NextHeader: &layers.UDP{
			DstPort: 53,
			SrcPort: 54321,
		},
	},
	{
		Description: "2 - A packet with IHL=6",
		IPv4Header: &layers.IPv4{
			Version: 4,
			IHL:     4,
			TTL:     64,
			Flags:   layers.IPv4DontFragment,
			SrcIP:   net.IPv4(4, 4, 4, 4),
			DstIP:   net.IPv4(1, 1, 1, 1),
		},
		NextHeader: &layers.UDP{
			DstPort: 53,
			SrcPort: 54321,
		},
	},
	{
		Description: "3 - A packet with IP PROTO=UDP but no UDP header",
		IPv4Header: &layers.IPv4{
			Version:  4,
			IHL:      5,
			TTL:      64,
			SrcIP:    net.IPv4(1, 2, 3, 4),
			DstIP:    net.IPv4(10, 20, 30, 40),
			Protocol: layers.IPProtocolUDP,
		},
		Size: 14 + 20,
	},
}

func TestMalformedPackets(t *testing.T) {
	RegisterTestingT(t)

	defer resetBPFMaps()

	for _, tc := range malformedTestCases {
		runBpfTest(t, "calico_from_host_ep", nil, func(bpfrun bpfProgRunFn) {
			_, _, _, _, _, pktBytes, err := generatePacket(ethDefault, tc.IPv4Header, nil, tc.NextHeader, tc.Payload, false)
			Expect(err).NotTo(HaveOccurred())
			if tc.Size != 0 {
				pktBytes = pktBytes[:tc.Size]
			}
			res, err := bpfrun(pktBytes)
			Expect(err).NotTo(HaveOccurred())
			pktR := gopacket.NewPacket(res.dataOut, layers.LayerTypeEthernet, gopacket.Default)
			fmt.Printf("pktR = %+v\n", pktR)
			Expect(res.RetvalStr()).To(Equal("TC_ACT_SHOT"), "expected the program to return TC_ACT_SHOT")
			Expect(res.dataOut).To(HaveLen(len(pktBytes)))
			Expect(res.dataOut).To(Equal(pktBytes))
		})
	}
}
