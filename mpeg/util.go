package mpeg

import "fmt"

type START_CODE_TYPE int

const (
	START_CODE_3 START_CODE_TYPE = 3
	START_CODE_4                 = 4
)

func FindStarCode(nalu []byte, offset int) (int, START_CODE_TYPE) {
	for i := offset; i < len(nalu)-4; i++ {
		if nalu[i] == 0x00 && nalu[i+1] == 0x00 {
			if nalu[i+2] == 0x01 {
				return i, START_CODE_3
			} else if nalu[i+2] == 0x00 && nalu[i+3] == 0x01 {
				return i, START_CODE_4
			}
		}
	}
	return -1, START_CODE_3
}

func FindSyncword(aac []byte, offset int) int {
	for i := offset; i < len(aac); i++ {
		if aac[i] == 0xFF && aac[i+1]&0xF0 == 0xF0 {
			return i
		}
	}
	return -1
}

func SplitFrame(frames []byte, onFrame func(nalu []byte) bool) {
	beg, sc := FindStarCode(frames, 0)
	for beg >= 0 {
		end, sc2 := FindStarCode(frames, beg+int(sc))
		if end == -1 {
			if onFrame != nil {
				onFrame(frames[beg+int(sc):])
			}
			break
		}
		if onFrame != nil && onFrame(frames[beg+int(sc):end]) == false {
			break
		}
		beg = end
		sc = sc2
	}
}

func SplitFrameWithStartCode(frames []byte, onFrame func(nalu []byte) bool) {
	beg, sc := FindStarCode(frames, 0)
	for {
		end, sc2 := FindStarCode(frames, beg+int(sc))
		if end == -1 {
			if onFrame != nil {
				onFrame(frames[beg:])
			}
			break
		}
		if onFrame != nil && onFrame(frames[beg:end]) == false {
			break
		}
		beg = end
		sc = sc2
	}
}

func SplitAACFrame(frames []byte, onFrame func(aac []byte)) {
	var adts ADTS_Frame_Header
	start := FindSyncword(frames, 0)
	for start >= 0 {
		adts.Decode(frames[start:])
		onFrame(frames[start : start+int(adts.Variable_Header.Frame_length)])
		start = FindSyncword(frames, start+int(adts.Variable_Header.Frame_length))
	}
}

func H264NaluType(h264 []byte) H264_NAL_TYPE {
	loc, sc := FindStarCode(h264, 0)
	return H264_NAL_TYPE(h264[loc+int(sc)] & 0x1F)
}

func H264NaluTypeWithoutStartCode(h264 []byte) H264_NAL_TYPE {
	return H264_NAL_TYPE(h264[0] & 0x1F)
}

func H265NaluType(h265 []byte) H265_NAL_TYPE {
	loc, sc := FindStarCode(h265, 0)
	return H265_NAL_TYPE((h265[loc+int(sc)] >> 1) & 0x3F)
}

func H265NaluTypeWithoutStartCode(h265 []byte) H265_NAL_TYPE {
	return H265_NAL_TYPE((h265[0] >> 1) & 0x3F)
}

func IsH264IDRFrame(h264 []byte) bool {

	ret := false
	onnalu := func(nalu []byte) bool {
		nal_type := H264NaluTypeWithoutStartCode(nalu)
		if nal_type < 5 {
			return false
		} else if nal_type == 5 {
			ret = true
			return false
		} else {
			return true
		}
	}
	SplitFrame(h264, onnalu)
	return ret
}

func IsH264VCLNaluType(nal_type H264_NAL_TYPE) bool {
	if nal_type <= H264_NAL_I_SLICE && nal_type > H264_NAL_RESERVED {
		return true
	}
	return false
}

func IsH265VCLNaluType(nal_type H265_NAL_TYPE) bool {
	if (nal_type <= H265_NAL_SLICE_CRA && nal_type >= H265_NAL_SLICE_BLA_W_LP) ||
		(nal_type <= H265_NAL_SLICE_RASL_R && nal_type >= H265_NAL_Slice_TRAIL_N) {
		return true
	}
	return false
}

func IsH265IDRFrame(h265 []byte) bool {
	ret := false
	onnalu := func(nalu []byte) bool {
		nal_type := H264NaluTypeWithoutStartCode(nalu)
		if nal_type <= 9 && nal_type >= 0 {
			return false
		} else if nal_type >= 16 && nal_type <= 21 {
			ret = true
			return false
		} else {
			return true
		}
	}
	SplitFrame(h265, onnalu)
	return ret
}

func ShowPacketHexdump(data []byte) {
	for k := 0; k < len(data); k++ {
		if k%8 == 0 && k != 0 {
			fmt.Printf("\n")
		}
		fmt.Printf("%02x ", data[k])
	}
	fmt.Printf("\n")
}

var crc32table [256]uint32 = [256]uint32{
	0x00000000, 0xB71DC104, 0x6E3B8209, 0xD926430D, 0xDC760413, 0x6B6BC517,
	0xB24D861A, 0x0550471E, 0xB8ED0826, 0x0FF0C922, 0xD6D68A2F, 0x61CB4B2B,
	0x649B0C35, 0xD386CD31, 0x0AA08E3C, 0xBDBD4F38, 0x70DB114C, 0xC7C6D048,
	0x1EE09345, 0xA9FD5241, 0xACAD155F, 0x1BB0D45B, 0xC2969756, 0x758B5652,
	0xC836196A, 0x7F2BD86E, 0xA60D9B63, 0x11105A67, 0x14401D79, 0xA35DDC7D,
	0x7A7B9F70, 0xCD665E74, 0xE0B62398, 0x57ABE29C, 0x8E8DA191, 0x39906095,
	0x3CC0278B, 0x8BDDE68F, 0x52FBA582, 0xE5E66486, 0x585B2BBE, 0xEF46EABA,
	0x3660A9B7, 0x817D68B3, 0x842D2FAD, 0x3330EEA9, 0xEA16ADA4, 0x5D0B6CA0,
	0x906D32D4, 0x2770F3D0, 0xFE56B0DD, 0x494B71D9, 0x4C1B36C7, 0xFB06F7C3,
	0x2220B4CE, 0x953D75CA, 0x28803AF2, 0x9F9DFBF6, 0x46BBB8FB, 0xF1A679FF,
	0xF4F63EE1, 0x43EBFFE5, 0x9ACDBCE8, 0x2DD07DEC, 0x77708634, 0xC06D4730,
	0x194B043D, 0xAE56C539, 0xAB068227, 0x1C1B4323, 0xC53D002E, 0x7220C12A,
	0xCF9D8E12, 0x78804F16, 0xA1A60C1B, 0x16BBCD1F, 0x13EB8A01, 0xA4F64B05,
	0x7DD00808, 0xCACDC90C, 0x07AB9778, 0xB0B6567C, 0x69901571, 0xDE8DD475,
	0xDBDD936B, 0x6CC0526F, 0xB5E61162, 0x02FBD066, 0xBF469F5E, 0x085B5E5A,
	0xD17D1D57, 0x6660DC53, 0x63309B4D, 0xD42D5A49, 0x0D0B1944, 0xBA16D840,
	0x97C6A5AC, 0x20DB64A8, 0xF9FD27A5, 0x4EE0E6A1, 0x4BB0A1BF, 0xFCAD60BB,
	0x258B23B6, 0x9296E2B2, 0x2F2BAD8A, 0x98366C8E, 0x41102F83, 0xF60DEE87,
	0xF35DA999, 0x4440689D, 0x9D662B90, 0x2A7BEA94, 0xE71DB4E0, 0x500075E4,
	0x892636E9, 0x3E3BF7ED, 0x3B6BB0F3, 0x8C7671F7, 0x555032FA, 0xE24DF3FE,
	0x5FF0BCC6, 0xE8ED7DC2, 0x31CB3ECF, 0x86D6FFCB, 0x8386B8D5, 0x349B79D1,
	0xEDBD3ADC, 0x5AA0FBD8, 0xEEE00C69, 0x59FDCD6D, 0x80DB8E60, 0x37C64F64,
	0x3296087A, 0x858BC97E, 0x5CAD8A73, 0xEBB04B77, 0x560D044F, 0xE110C54B,
	0x38368646, 0x8F2B4742, 0x8A7B005C, 0x3D66C158, 0xE4408255, 0x535D4351,
	0x9E3B1D25, 0x2926DC21, 0xF0009F2C, 0x471D5E28, 0x424D1936, 0xF550D832,
	0x2C769B3F, 0x9B6B5A3B, 0x26D61503, 0x91CBD407, 0x48ED970A, 0xFFF0560E,
	0xFAA01110, 0x4DBDD014, 0x949B9319, 0x2386521D, 0x0E562FF1, 0xB94BEEF5,
	0x606DADF8, 0xD7706CFC, 0xD2202BE2, 0x653DEAE6, 0xBC1BA9EB, 0x0B0668EF,
	0xB6BB27D7, 0x01A6E6D3, 0xD880A5DE, 0x6F9D64DA, 0x6ACD23C4, 0xDDD0E2C0,
	0x04F6A1CD, 0xB3EB60C9, 0x7E8D3EBD, 0xC990FFB9, 0x10B6BCB4, 0xA7AB7DB0,
	0xA2FB3AAE, 0x15E6FBAA, 0xCCC0B8A7, 0x7BDD79A3, 0xC660369B, 0x717DF79F,
	0xA85BB492, 0x1F467596, 0x1A163288, 0xAD0BF38C, 0x742DB081, 0xC3307185,
	0x99908A5D, 0x2E8D4B59, 0xF7AB0854, 0x40B6C950, 0x45E68E4E, 0xF2FB4F4A,
	0x2BDD0C47, 0x9CC0CD43, 0x217D827B, 0x9660437F, 0x4F460072, 0xF85BC176,
	0xFD0B8668, 0x4A16476C, 0x93300461, 0x242DC565, 0xE94B9B11, 0x5E565A15,
	0x87701918, 0x306DD81C, 0x353D9F02, 0x82205E06, 0x5B061D0B, 0xEC1BDC0F,
	0x51A69337, 0xE6BB5233, 0x3F9D113E, 0x8880D03A, 0x8DD09724, 0x3ACD5620,
	0xE3EB152D, 0x54F6D429, 0x7926A9C5, 0xCE3B68C1, 0x171D2BCC, 0xA000EAC8,
	0xA550ADD6, 0x124D6CD2, 0xCB6B2FDF, 0x7C76EEDB, 0xC1CBA1E3, 0x76D660E7,
	0xAFF023EA, 0x18EDE2EE, 0x1DBDA5F0, 0xAAA064F4, 0x738627F9, 0xC49BE6FD,
	0x09FDB889, 0xBEE0798D, 0x67C63A80, 0xD0DBFB84, 0xD58BBC9A, 0x62967D9E,
	0xBBB03E93, 0x0CADFF97, 0xB110B0AF, 0x060D71AB, 0xDF2B32A6, 0x6836F3A2,
	0x6D66B4BC, 0xDA7B75B8, 0x035D36B5, 0xB440F7B1,
}

func CalcCrc32(crc uint32, buffer []byte) uint32 {
	var i int = 0
	for i = 0; i < len(buffer); i++ {
		crc = crc32table[(crc^uint32(buffer[i]))&0xff] ^ (crc >> 8)
	}
	return crc
}
