package core

import (
	"fmt"
	"time"

	"github.com/aler9/gortsplib/v2/pkg/format"
	"github.com/aler9/gortsplib/v2/pkg/formatdecenc/rtpvp9"
	"github.com/pion/rtp"
)

type dataVP9 struct {
	rtpPackets []*rtp.Packet
	ntp        time.Time
	pts        time.Duration
	frame      []byte
}

func (d *dataVP9) getRTPPackets() []*rtp.Packet {
	return d.rtpPackets
}

func (d *dataVP9) getNTP() time.Time {
	return d.ntp
}

type formatProcessorVP9 struct {
	format  *format.VP9
	encoder *rtpvp9.Encoder
	decoder *rtpvp9.Decoder
}

func newFormatProcessorVP9(
	forma *format.VP9,
	allocateEncoder bool,
) (*formatProcessorVP9, error) {
	t := &formatProcessorVP9{
		format: forma,
	}

	if allocateEncoder {
		t.encoder = forma.CreateEncoder()
	}

	return t, nil
}

func (t *formatProcessorVP9) process(dat data, hasNonRTSPReaders bool) error { //nolint:dupl
	tdata := dat.(*dataVP9)

	if tdata.rtpPackets != nil {
		pkt := tdata.rtpPackets[0]

		// remove padding
		pkt.Header.Padding = false
		pkt.PaddingSize = 0

		if pkt.MarshalSize() > maxPacketSize {
			return fmt.Errorf("payload size (%d) is greater than maximum allowed (%d)",
				pkt.MarshalSize(), maxPacketSize)
		}

		// decode from RTP
		if hasNonRTSPReaders {
			if t.decoder == nil {
				t.decoder = t.format.CreateDecoder()
			}

			frame, pts, err := t.decoder.Decode(pkt)
			if err != nil {
				if err == rtpvp9.ErrMorePacketsNeeded {
					return nil
				}
				return err
			}

			tdata.frame = frame
			tdata.pts = pts
		}

		// route packet as is
		return nil
	}

	pkts, err := t.encoder.Encode(tdata.frame, tdata.pts)
	if err != nil {
		return err
	}

	tdata.rtpPackets = pkts
	return nil
}
